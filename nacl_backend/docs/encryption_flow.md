# Encryption Flow - NACL Backend

This document describes the complete encryption/decryption flow for the NACL password manager.

---

## Table of Contents

- [Overview](#overview)
- [Key Concepts](#key-concepts)
- [Two Types of Passwords](#two-types-of-passwords)
- [Registration Flow](#registration-flow)
- [Add Service Password Flow](#add-service-password-flow)
- [Retrieve Service Password Flow](#retrieve-service-password-flow)
- [Change Login Password Flow](#change-login-password-flow)
- [Security Considerations](#security-considerations)

---

## Overview

NACL uses a **two-key encryption system**:

1. **Login Password** - User's master password for authentication and key derivation
2. **Master Key** - Random 32-byte key for encrypting/decrypting service passwords
3. **Service Passwords** - Actual passwords for Gmail, Facebook, etc. (encrypted with master key)

**Key Design Decision:** The login password is **never used directly** to encrypt service passwords. Instead, it encrypts a randomly generated master key, which then encrypts all service passwords.

**Benefits:**
- Changing login password only requires re-encrypting the master key (not all service passwords)
- Master key is cryptographically random (stronger than user-chosen passwords)
- Clear separation between authentication and encryption

---

## Key Concepts

### Salt

**What:** Random 32-byte value generated during registration  
**Purpose:** Ensures unique key derivation even if two users have the same password  
**Storage:** Plaintext in `users.masterKeySalt`  
**Security:** Salt is NOT secret - its purpose is uniqueness, not confidentiality

### Key Derivation

**What:** Process of converting password + salt → fixed-length encryption key  
**Algorithm:** Argon2id (memory-hard, GPU/ASIC resistant)  
**Output:** 32-byte derived key  
**Formula:** `derivedKey = Argon2id(loginPassword, salt)`

### Master Key

**What:** Random 32-byte data encryption key  
**Purpose:** Encrypts/decrypts all service passwords for a user  
**Storage:** Encrypted with derived key in `users.encryptedMasterKey`  
**Security:** NEVER stored in plaintext, NEVER logged, transmitted only over internal trusted network

### Nonce

**What:** Random value used once per encryption operation  
**Purpose:** Ensures same plaintext encrypts to different ciphertext each time  
**Size:** 12 bytes (96 bits) for AES-GCM  
**Storage:** Plaintext in `ciphers.nonce`  
**Security:** Must be unique per encryption, but NOT secret

---

## Base64 Encoding Boundary

The encryption functions (`DeriveKey`, `Encrypt`, `Decrypt`) operate exclusively on raw bytes (`[]byte`). PostgreSQL text columns store strings. This creates a **boundary** where data must be converted between `[]byte` and `string`.

### The Rule

| Direction | Function | Example |
|-----------|----------|---------|
| **Storage** (raw bytes → string) | `base64.StdEncoding.EncodeToString()` | `encodedSalt := base64.StdEncoding.EncodeToString(salt)` |
| **Retrieval** (string → raw bytes) | `base64.StdEncoding.DecodeString()` | `salt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)` |

### Fields Affected

| Database Column | Stored As | Decode Before Use |
|----------------|-----------|-------------------|
| `users.masterKeySalt` | Base64 string | `base64.StdEncoding.DecodeString()` → pass to `DeriveKey()` |
| `users.encryptedMasterKey` | Base64 string | `base64.StdEncoding.DecodeString()` → pass to `Decrypt()` |

### Common Mistake

Never convert a base64 string to bytes by casting:

```go
// WRONG: gives ASCII bytes of the base64 string, not the original data
salt := []byte(user.MasterKeySalt)  // []byte("MWkwRjRW...") — wrong!

// CORRECT: decodes the base64 string back to the original raw bytes
salt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)  // 32 bytes
```

Casting `[]byte(string)` on a base64-encoded value produces the wrong length and wrong data. `DeriveKey` requires exactly 32 bytes for the salt — a base64 string cast to `[]byte` will be longer and will fail with "length of salt must be 32 bytes".

---

## Two Types of Passwords

### Login Password (Master Password)

| Property | Value |
|----------|-------|
| **Purpose** | Authenticate user + derive encryption key |
| **Provided by** | User at registration and login |
| **Stored as** | Bcrypt hash in `users.passwordHash` |
| **Example** | `"MyMasterPassword123!"` |
| **Frequency** | Entered once per session |
| **Used for** | - Account login<br>- Deriving key to decrypt master key |

### Service Password

| Property | Value |
|----------|-------|
| **Purpose** | Actual password for external services |
| **Provided by** | User when adding a service |
| **Stored as** | Encrypted ciphertext in `ciphers.encryptedPassword` |
| **Example** | `"GmailPassword456!"`, `"FacebookPwd789!"` |
| **Frequency** | One per service (Gmail, Twitter, etc.) |
| **Used for** | Logging into Gmail, Facebook, etc. |

---

## Registration Flow

### User Input

```
username: "john@example.com"
name: "John"
loginPassword: "MyMasterPassword123!"
```

### System Processing

```
┌─────────────────────────────────────────────────────────────┐
│                    REGISTRATION                             │
└─────────────────────────────────────────────────────────────┘

Step 1: Generate random values
  ├─ salt = crypto.randomBytes(32)
  │   → "a3f5b8c2d9e1f4a7b0c3d6e9f2a5b8c1d4e7f0a3b6c9d2e5f8a1b4c7d0e3f6"
  │
  └─ masterKey = crypto.randomBytes(32)
      → "7f2a9b4c6d8e1f3a5b7c9d0e2f4a6b8c0d2e4f6a8b0c2d4e6f8a0b2c4d6e8f0a2"

Step 2: Derive encryption key from login password
  ├─ derivedKey = Argon2id(loginPassword, salt)
  │   → 32-byte key derived from password + salt
  │
  └─ passwordHash = bcrypt(loginPassword)
      → "$2b$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

Step 3: Encrypt master key with derived key
  └─ encryptedMasterKey = AES-GCM-encrypt(masterKey, derivedKey)
      → "base64-encoded-ciphertext-with-auth-tag"

Step 4: Store in database
  └─ INSERT INTO users (
       username,
       name,
       passwordHash,
       masterKeySalt,
       encryptedMasterKey
     )
```

### Database State After Registration

```sql
SELECT * FROM users WHERE username = 'john@example.com';

-- Result:
-- id: uuid
-- username: "john@example.com"
-- name: "John"
-- passwordHash: "$2b$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
-- masterKeySalt: "a3f5b8c2d9e1f4a7b0c3d6e9f2a5b8c1d4e7f0a3b6c9d2e5f8a1b4c7d0e3f6"
-- encryptedMasterKey: "base64-ciphertext"
-- createdAt: 2026-05-23T12:00:00Z
-- updatedAt: 2026-05-23T12:00:00Z
```

---

## Add Service Password Flow

### User Input

```
service: "Gmail"
serviceUsername: "john@gmail.com"
servicePassword: "GmailPassword456!"
loginPassword: "MyMasterPassword123!"  -- To authorize encryption
```

### System Processing

```
┌─────────────────────────────────────────────────────────────┐
│              ADD SERVICE PASSWORD (Gmail)                   │
└─────────────────────────────────────────────────────────────┘

Step 1: Fetch user from database
  └─ SELECT * FROM users WHERE id = :userId
      → Returns: masterKeySalt (base64 string), encryptedMasterKey (base64 string)

Step 2: Decode base64 fields for cryptographic use
  ├─ decodedSalt = base64.StdEncoding.DecodeString(masterKeySalt)
  └─ decodedMasterKey = base64.StdEncoding.DecodeString(encryptedMasterKey)

Step 3: Derive encryption key from login password
  └─ derivedKey = Argon2id(loginPassword, decodedSalt)
      → Same derived key as registration (same password + salt)

Step 3: Decrypt master key with derived key
  └─ masterKey = AES-GCM-decrypt(decodedMasterKey, derivedKey)
      → Recovers original 32-byte random master key

Step 4: Encrypt service password with master key
  ├─ Generates random nonce (12 bytes)
  ├─ Encrypts: ciphertext = AES-GCM-encrypt(servicePassword, masterKey, nonce)
  └─ Returns: { encrypted: "base64-ciphertext", nonce: "base64-nonce" }

Step 5: Store encrypted service password in database
  └─ INSERT INTO ciphers (
       service,
       serviceUsername,
       description,
       encryptedPassword,
       nonce,
       encryptionAlgorithm,
       userId
     )
```

### Database State After Adding Service Password

```sql
SELECT * FROM ciphers WHERE userId = :userId;

-- Result:
-- id: uuid
-- service: "Gmail"
-- serviceUsername: "john@gmail.com"
-- description: "Personal Gmail"
-- encryptedPassword: "base64-ciphertext"
-- nonce: "base64-nonce"
-- encryptionAlgorithm: "aes-256-gcm"
-- userId: uuid
-- createdAt: 2026-05-23T12:05:00Z
-- updatedAt: 2026-05-23T12:05:00Z
-- lastUsedAt: NULL
```

---

## Retrieve Service Password Flow

### User Input

```
serviceId: "uuid-of-gmail-cipher"
loginPassword: "MyMasterPassword123!"  -- To authorize decryption
```

### System Processing

```
┌─────────────────────────────────────────────────────────────┐
│           RETRIEVE SERVICE PASSWORD (Gmail)                 │
└─────────────────────────────────────────────────────────────┘

Step 1: Fetch user from database
  └─ SELECT * FROM users WHERE id = :userId
      → Returns: masterKeySalt (base64 string), encryptedMasterKey (base64 string)

Step 2: Fetch cipher from database
  └─ SELECT * FROM ciphers WHERE id = :serviceId AND userId = :userId
      → Returns: encryptedPassword (base64 string), nonce (base64 string), encryptionAlgorithm

Step 3: Decode base64 fields for cryptographic use
  ├─ decodedSalt = base64.StdEncoding.DecodeString(masterKeySalt)
  ├─ decodedMasterKey = base64.StdEncoding.DecodeString(encryptedMasterKey)
  └─ decodedPassword = base64.StdEncoding.DecodeString(encryptedPassword)

Step 4: Derive encryption key from login password
  └─ derivedKey = Argon2id(loginPassword, decodedSalt)
      → Same derived key as before

Step 5: Decrypt master key with derived key
  └─ masterKey = AES-GCM-decrypt(decodedMasterKey, derivedKey)
      → Recovers original 32-byte random master key

Step 6: Decrypt service password
  └─ plaintext = AES-GCM-decrypt(decodedPassword, masterKey)
      → Returns: "GmailPassword456!"

Step 6: Update lastUsedAt timestamp
  └─ UPDATE ciphers SET lastUsedAt = NOW() WHERE id = :serviceId

Step 7: Return decrypted password to user
  └─ Return: "GmailPassword456!"
```

### Security Notes for Retrieval

- **Rate limiting:** Maximum 5 decryption attempts per minute per user
- **Audit logging:** Log timestamp, service name, IP address (NOT the password)
- **Memory handling:** Derived key and master key are cleared from memory after use
- **HTTPS required:** All decryption requests must use encrypted transport

---

## Change Login Password Flow

### User Input

```
oldPassword: "MyMasterPassword123!"
newPassword: "NewMasterPassword456!"
```

### System Processing

```
┌─────────────────────────────────────────────────────────────┐
│                    CHANGE PASSWORD                          │
└─────────────────────────────────────────────────────────────┘

Step 1: Verify old password
  └─ bcrypt.verify(oldPassword, passwordHash)
      → Must return true

Step 2: Derive old encryption key
  └─ decodedSalt = base64.StdEncoding.DecodeString(masterKeySalt)
  └─ oldDerivedKey = Argon2id(oldPassword, decodedSalt)

Step 3: Decrypt master key with old derived key
  └─ decodedMasterKey = base64.StdEncoding.DecodeString(encryptedMasterKey)
  └─ masterKey = AES-GCM-decrypt(decodedMasterKey, oldDerivedKey)
      → Recovers original 32-byte random master key
      → ✅ Ciphers remain encrypted with this same master key!

Step 4: Generate new salt (rotate salt for security)
  └─ newSalt = crypto.randomBytes(32)
      → "b4g6c9d2e5f8a1b4c7d0e3f6a9b2c5d8e1f4a7b0c3d6e9f2a5b8c1d4e7f0a3b6"

Step 5: Derive new encryption key from new password
  └─ newDerivedKey = Argon2id(newPassword, newSalt)

Step 6: Re-encrypt master key with new derived key
  └─ newEncryptedMasterKey = AES-GCM-encrypt(masterKey, newDerivedKey)

Step 7: Hash new password
  └─ newPasswordHash = bcrypt.hash(newPassword)

Step 8: Update user record
  └─ UPDATE users SET
       passwordHash = newPasswordHash,
       masterKeySalt = newSalt,
       encryptedMasterKey = newEncryptedMasterKey,
       updatedAt = NOW()
     WHERE id = :userId

✅ IMPORTANT: No cipher records are modified!
   - encryptedPassword stays the same
   - nonce stays the same
   - Master key stays the same
   - Only the wrapper (encryptedMasterKey) changes
```

### Why This Design?

| Old Approach (Option A) | New Approach (Option B) |
|-------------------------|-------------------------|
| Decrypt ALL ciphers with old key | Decrypt masterKey once |
| Re-encrypt ALL ciphers with new key | Re-encrypt masterKey once |
| Slow for users with many passwords | Fast regardless of cipher count |
| Risk of data loss if process interrupted | Atomic operation, safe |
| ❌ Don't use this | ✅ Current implementation |

---

## Security Considerations

### ✅ Implemented Security Measures

| Measure | Implementation |
|---------|----------------|
| **Separate login and encryption** | Login password ≠ master key |
| **Random master key** | 32 bytes from CSPRNG |
| **Per-cipher nonce** | Unique 12-byte nonce per encryption |
| **Salt for key derivation** | Random 32-byte salt per user |
| **Strong KDF** | Argon2id (memory-hard) |
| **Authenticated encryption** | AES-GCM with auth tag |
| **Cascade delete** | Deleting user removes all ciphers |
| **Password hashing** | Bcrypt for login password |

### ⚠️ Security Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Salt stored plaintext | Attacker has salt | Salt is useless without password |
| Master key in memory during operation | Brief exposure | Cleared immediately after use |
| HTTPS required | TLS overhead | Use TLS 1.3, HTTP/2 |
| Ciphertext length reveals plaintext length | Attacker can infer password length | Acceptable for password manager use case; length doesn't reveal content or significantly reduce keyspace |

### 🔐 Ciphertext Length Decision

**Decision:** Do not pad ciphertext to fixed length (MVP)

**Rationale:**
- Ciphertext length = plaintext length + 28 bytes (nonce + auth tag)
- After base64 encoding: ~1.33x overhead
- Example: 32-byte master key → 60 bytes encrypted → ~80 chars base64

**Why This Is Acceptable:**
1. **Password length already known** - Users know how many passwords they store and roughly how long they are
2. **Length doesn't reveal content** - Knowing a password is 12 characters doesn't help crack it (could be weak or strong)
3. **Minimal security benefit** - Compared to other threats (nonce reuse, weak keys, access control)
4. **Significant trade-offs** - Padding wastes storage, adds complexity, false sense of security

**When Length Leakage Matters:**
- ✅ Encrypting documents (1 page vs 100 pages reveals sensitive info)
- ✅ Encrypting messages ("yes" vs detailed response)
- ✅ Encrypting search queries (short vs long reveals intent)
- ❌ Passwords (not a significant concern for password managers)

**Future Enhancement (Post-MVP):**
If length hiding becomes important, implement **bucketed padding**:
```go
// Round up to nearest 32 bytes
func padToBucket(data []byte) []byte {
    bucket := 32
    remainder := len(data) % bucket
    if remainder == 0 {
        return data
    }
    padding := bucket - remainder
    return append(data, make([]byte, padding)...)
}
```

This hides exact length while being more efficient than fixed-length padding.

**Industry Standard:** Most password managers (1Password, Bitwarden, LastPass) do not pad ciphertext length.

### 🔐 Master Key Handling

**Security Controls:**

| Control | Implementation |
|---------|----------------|
| **No persistence** | Master key exists only in memory during the operation |
| **No logging** | MUST NOT log keys, passwords, or plaintext data |
| **Memory clearing** | Zero out buffers immediately after encryption/decryption |
| **TLS 1.3 minimum** | All communication encrypted |

**What This Protects Against:**
- ✅ External attackers
- ✅ Network sniffing (TLS encryption)

**What This Does NOT Protect Against:**
- ⚠️ Compromised server (has access to all master keys)
- ⚠️ Compromised host (memory inspection attacks)

**Mitigation Strategies:**
- Keep codebase minimal (small attack surface, single responsibility)
- Regular security audits
- Run as non-root user
- Use Docker security best practices (read-only filesystem, no capabilities)

### 🔒 Recommended Security Enhancements

1. **Rate Limiting**
   - Limit decryption attempts to 5 per minute per user
   - Prevents brute force attacks on decryption endpoint

2. **Audit Logging**
   - Log password access events (timestamp, service name, IP address)
   - Do NOT log the actual passwords
   - Track user activity for security monitoring

3. **Session Management**
   - Use short-lived JWT tokens (15 minutes)
   - Implement refresh tokens for extended sessions
   - Invalidate all sessions on password change

4. **Memory Security**
   - Clear sensitive data (derived keys, master keys) from memory after use
   - Zero out buffers containing cryptographic material
   - Avoid logging sensitive values

### 🚫 What NOT to Do

- ❌ **Never log encrypted passwords, keys, or plaintext passwords**
- ❌ **Never store plaintext passwords**
- ❌ **Never transmit master key over PUBLIC network** (only internal trusted network)
- ❌ **Never reuse nonces** (compromises AES-GCM security)
- ❌ **Never use ECB mode** (use GCM or CBC with HMAC)
- ❌ **Never use MD5 or SHA1** for key derivation (use Argon2id)
- ❌ **Never skip authentication tag verification** (allows tampering)

---

## Glossary

| Term | Definition |
|------|------------|
| **Login Password** | User's master password for account access |
| **Service Password** | Password for external services (Gmail, etc.) |
| **Master Key** | Random 32-byte key for encrypting service passwords |
| **Salt** | Random value for unique key derivation |
| **Nonce** | Random value used once per encryption |
| **Derived Key** | Output of Argon2id(password + salt) |
| **KDF** | Key Derivation Function (Argon2id) |
| **AES-GCM** | Authenticated encryption algorithm |
| **Ciphertext** | Encrypted data |
| **Plaintext** | Unencrypted data |

---

## References

- [Argon2 RFC 9106](https://www.rfc-editor.org/rfc/rfc9106.html)
- [AES-GCM Specification](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [NIST SP 800-132](https://csrc.nist.gov/publications/detail/sp/800-132/final) - Password-Based Key Derivation
