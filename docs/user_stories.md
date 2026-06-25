# User Stories - NaCl Password Manager

## Overview
This document tracks all user stories, their requirements, and task breakdowns for the NaCl password manager.

## Navigation & Architecture
- **Single Page App** with React Router for client-side routing
- **Top tab navigation** with 4 main sections: Save new, My passwords, Log, User
- **Plain CSS** for styling (no UI framework)
- **Auth-protected routes** - all pages require authentication except login/register

---

## Development Approach

### Testing Strategy
- **Backend-first TDD** - Each component developed with unit/integration tests
- **Encryption foundation** - Tested in isolation first
- **Database migrations** - All at once (schemas known upfront)
- **Frontend after backend** - E2E tests once FE is complete
- **TDD for backend** - Unit tests for encryption, integration tests for API endpoints + DB

### Task Numbering Convention
- `US-X.Y-B` = User Story X, Task Y, **Backend**
- `US-X.Y-F` = User Story X, Task Y, **Frontend**
- `US-X.Y-NH` = User Story X, Task Y, **Nice-to-have**

---

## Priority Order & Implementation Sequence

### Phase 0: Database & Encryption Foundation (Prerequisites)
**Must be completed before any feature development**

1. **US-0: Database Migrations** - All tables (users, ciphers, operations)
2. **US-1: Encryption Foundation** - AES-256-GCM + Argon2id implementation

### Phase 1: Core Backend (TDD Approach)
**Backend tasks only - frontend waits until Phase 4**

3. **US-2: User Registration** - POST /api/users, master key generation
4. **US-3: User Login** - JWT authentication
5. **US-4: User Logout** - Client-side logout (minimal backend)

### Phase 2: Password Management Backend
**Backend tasks only - core product functionality**

6. **US-5: Save Credential Passwords** - Encrypt + store
7. **US-6: View Passwords List** - Full CRUD + decrypt endpoints
8. **US-7: Change Login Password** - Key rotation
9. **US-8: Audit Trail** - Operations logging

### Phase 3: Frontend Foundation (Granular Breakdown)
**Infrastructure before features**

10. **US-9a: React Router Setup** - Basic routing configuration
11. **US-9b: Main Layout** - Header + nav tabs + content area
12. **US-9c: Auth Guard** - Route protection wrapper
13. **US-9d: Common Components** - Loading, toast, empty state

### Phase 4: Frontend Features
**All frontend feature components**

14. **US-2-F through US-8-F** - All frontend features (registration through audit log)

### Phase 5: E2E Testing
**End-to-end tests for complete user flows**

15. **E2E Test Suite** - Complete user journey testing

---

## User Stories

### US-0: Database Migrations
**As a** developer  
**I want to** create all database tables upfront  
**So that** we have a stable schema for backend development

**Acceptance Criteria:**
- Separate migration files for each table (easy to rollback/modify)
- Users table with master key fields
- Ciphers table for encrypted service passwords
- Operations table for audit trail
- All necessary indexes and constraints

**Tasks:**
- [ ] US-0.1-B: Migration 00001_users.sql (users table with masterKeySalt, encryptedMasterKey)
- [ ] US-0.2-B: Migration 00002_ciphers.sql (ciphers table for encrypted passwords)
- [ ] US-0.3-B: Migration 00003_operations.sql (operations table for audit trail)
- [ ] US-0.4-B: sqlc queries generation for all tables
- [ ] US-0.5-B: Database models generation (db package)

**Priority:** 1 (Phase 0)

---

### US-1: Encryption Foundation
**As a** developer  
**I want to** implement encryption/decryption functions  
**So that** we can securely encrypt service passwords

**Acceptance Criteria:**
- Standalone package (`internal/crypto/`)
- AES-256-GCM encryption/decryption
- Argon2id key derivation
- Unit tested in isolation
- Clean API for other packages to use

**Tasks:**
- [ ] US-1.1-B: Create internal/crypto package structure
- [ ] US-1.2-B: Implement AES-256-GCM encrypt function
- [ ] US-1.3-B: Implement AES-256-GCM decrypt function
- [ ] US-1.4-B: Implement Argon2id key derivation function
- [ ] US-1.5-B: Implement master key generation (random 32-byte)
- [ ] US-1.6-B: Implement salt generation (random 32-byte)
- [ ] US-1.7-B: Unit tests for encryption/decryption round-trip
- [ ] US-1.8-B: Unit tests for key derivation
- [ ] US-1.9-B: Unit tests for master key/salt generation

**Priority:** 2 (Phase 0)

---

### US-2: User Registration
**As a** user  
**I want to** register using a username and password  
**So that** I can create an account

**Acceptance Criteria:**
- Username: free-form text (any characters allowed)
- Password: free-form text (no complexity requirements)
- Password field hidden when typing (password input type)
- Visual feedback for: success, validation errors, server errors
- After successful registration → redirect to login page
- Generate and store master key + salt during registration
- No rate limiting, CAPTCHA, or email verification (MVP scope)

**Tasks:**
- [ ] US-2.1-B: POST /api/users endpoint *(already implemented)*
- [ ] US-2.2-B: Username/password validation logic *(partially done)*
- [ ] US-2.3-B: Argon2id password hashing *(already implemented)*
- [ ] US-2.4-B: Generate random 32-byte master key during registration
- [ ] US-2.5-B: Generate random 32-byte salt during registration
- [ ] US-2.6-B: Encrypt master key with derived key (Argon2id)
- [ ] US-2.7-B: Store masterKeySalt and encryptedMasterKey in users table
- [ ] US-2.8-B: Integration tests for registration endpoint
- [ ] US-2.1-F: Registration form component (username + password fields)
- [ ] US-2.2-F: Password input with hidden text
- [ ] US-2.3-F: Form validation (client-side)
- [ ] US-2.4-F: Success/error message display
- [ ] US-2.5-F: Redirect to login page on success
- [ ] US-2.6-F: Frontend service to call registration endpoint

**Nice-to-have:**
- [ ] US-2.7-NH: Rate limiting on registration endpoint
- [ ] US-2.8-NH: CAPTCHA verification
- [ ] US-2.9-NH: Email verification

**Priority:** 3 (Phase 1)

---

### US-3: User Login
**As a** user  
**I want to** login using my username and password  
**So that** I can access my account

**Acceptance Criteria:**
- Login form with username + password fields
- JWT token authentication
- Token stored in localStorage
- Generic error message for failed login ("Invalid credentials")
- Visual feedback for: invalid credentials, server errors
- After successful login → redirect to main page
- Hardcoded session timeout (custom timeout as nice-to-have)

**Tasks:**
- [ ] US-3.1-B: POST /api/auth/login endpoint
- [ ] US-3.2-B: Password verification (compare with stored hash)
- [ ] US-3.3-B: JWT token generation
- [ ] US-3.4-B: JWT token expiration (hardcoded value)
- [ ] US-3.5-B: Auth middleware for protected routes
- [ ] US-3.6-B: Integration tests for login endpoint
- [ ] US-3.1-F: Login form component
- [ ] US-3.2-F: Token storage in localStorage
- [ ] US-3.3-F: Auth state management
- [ ] US-3.4-F: Frontend service to call login endpoint
- [ ] US-3.5-F: Protected route handling (redirect if not authenticated)

**Nice-to-have:**
- [ ] US-3.6-NH: User-configurable session timeout
- [ ] US-3.7-NH: Account lockout after N failed attempts
- [ ] US-3.8-NH: "Remember me" checkbox

**Priority:** 4 (Phase 1)

---

### US-4: User Logout
**As a** user  
**I want to** log out of my session  
**So that** I can securely end my authenticated session

**Acceptance Criteria:**
- Client-side logout (remove token from localStorage)
- JWT invalidation: client-side only (token valid until natural expiration)
- Clear all auth state from client
- Redirect to login page after logout
- Visual confirmation of logout

**Tasks:**
- [ ] US-4.1-B: No backend changes required for MVP (client-side logout)
- [ ] US-4.1-F: Logout button/component
- [ ] US-4.2-F: Clear token from localStorage
- [ ] US-4.3-F: Clear auth state
- [ ] US-4.4-F: Redirect to login page
- [ ] US-4.5-F: Visual confirmation message

**Nice-to-have:**
- [ ] US-4.6-NH: Token invalidation endpoint with blacklist

**Priority:** 5 (Phase 1)

---

### US-5: Save Credential Passwords
**As a** user  
**I want to** save passwords and account names for services  
**So that** I can securely store my credentials encrypted

**Acceptance Criteria:**
- Form fields: service name, service username, service password (with confirmation), optional notes/description
- Only AES-256-GCM encryption (MVP)
- Monolith architecture (encryption library in same process)
- User must re-enter login password to authorize each encryption operation
- Database schema changes tracked as separate tasks

**Tasks:**
- [ ] US-5.1-B: POST /api/ciphers endpoint (create encrypted password)
- [ ] US-5.2-B: Login password verification for encryption authorization
- [ ] US-5.3-B: Integration tests for create cipher endpoint
- [ ] US-5.1-F: Frontend form component (service name, username, password + confirmation, notes)
- [ ] US-5.2-F: Frontend service to call encryption endpoint
- [ ] US-5.3-F: Password confirmation validation (client-side)

**Priority:** 6 (Phase 2)

---

### US-6: View Saved Passwords List
**As a** user  
**I want to** see my saved passwords as a list showing service name, username, and notes  
**So that** I can browse, view, edit, and delete my credentials

**Acceptance Criteria:**
- List shows: service name, service username, notes
- Passwords hidden completely in list view
- Simple list with natural scrolling (no pagination for MVP)
- Search/filter input (by service name or username) - frontend filtering
- Each row has:
  - Decrypt button - Shows masked password (••••••••)
  - Reveal button - Shows actual plaintext password (after decryption)
  - Update/Edit button - Opens edit form (modal or inline)
  - Delete button - Remove password (with confirmation)
- Empty state message when no saved passwords
- Loading state while fetching/decrypting

**Tasks:**
- [ ] US-6.1-B: GET /api/ciphers endpoint (return ALL user's passwords, metadata only)
- [ ] US-6.2-B: GET /api/ciphers/:id/decrypt endpoint (decrypt single password)
- [ ] US-6.3-B: PUT /api/ciphers/:id endpoint (update encrypted password)
- [ ] US-6.4-B: DELETE /api/ciphers/:id endpoint (delete password)
- [ ] US-6.5-B: Login password verification for encryption/decryption authorization
- [ ] US-6.6-B: Integration tests for all cipher endpoints
- [ ] US-6.1-F: Password list component
- [ ] US-6.2-F: List item component (service row/card)
- [ ] US-6.3-F: Search/filter input component (frontend filtering by service or username)
- [ ] US-6.4-F: Empty state component
- [ ] US-6.5-F: Frontend service to fetch passwords list
- [ ] US-6.6-F: Decrypt button + masked password display
- [ ] US-6.7-F: Reveal button + plaintext password display
- [ ] US-6.8-F: Update/Edit form (modal or inline)
- [ ] US-6.9-F: Delete button (with confirmation dialog)
- [ ] US-6.10-F: Loading state
- [ ] US-6.11-F: Client-side filter logic

**Priority:** 7 (Phase 2)

---

### US-7: Change Login Password
**As a** user  
**I want to** change my login password  
**So that** I can update my master password while maintaining access to all my saved service passwords

**Acceptance Criteria:**
- Form fields: old password, new password, confirm new password
- Free-form password (no complexity requirements)
- No password strength indicator (MVP)
- Visual feedback for: wrong old password, mismatch confirmation, server errors
- After successful change → force re-login (logout all sessions)

**Encryption Flow:**
1. Verify old password (bcrypt)
2. Derive old key (Argon2id with old password + existing salt)
3. Decrypt master key with old derived key
4. Generate new salt (rotate for security)
5. Derive new key (Argon2id with new password + new salt)
6. Re-encrypt master key with new derived key
7. Hash new password (bcrypt)
8. Atomic database update (all or nothing)

**Tasks:**
- [ ] US-7.1-B: PUT /api/users/password endpoint
- [ ] US-7.2-B: Old password verification (bcrypt)
- [ ] US-7.3-B: Master key decryption with old derived key
- [ ] US-7.4-B: New salt generation (random 32-byte)
- [ ] US-7.5-B: New derived key derivation (Argon2id)
- [ ] US-7.6-B: Master key re-encryption with new derived key
- [ ] US-7.7-B: New password hashing (bcrypt)
- [ ] US-7.8-B: Atomic database update (transaction)
- [ ] US-7.9-B: Integration tests for password change endpoint
- [ ] US-7.1-F: Change password form component
- [ ] US-7.2-F: Old password input
- [ ] US-7.3-F: New password input (with confirmation)
- [ ] US-7.4-F: Validation logic (match confirmation)
- [ ] US-7.5-F: Success/error messaging
- [ ] US-7.6-F: Force logout + redirect to login page

**Nice-to-have:**
- [ ] US-7.7-NH: Rate limiting on password change
- [ ] US-7.8-NH: Audit log entry
- [ ] US-7.9-NH: Email notification

**Priority:** 8 (Phase 2)

---

### US-8: Password Update History (Audit Trail)
**As a** user  
**I want to** see a trace of all password operations (create, update, delete) with date and optional notes  
**So that** I can track the history of my credential changes for security auditing

**Acceptance Criteria:**
- Track all cipher operations: create, update, delete
- New `operations` table with: timestamp, optional notes, service name, action type
- Separate global log page (not per-password view)
- Filter input (by service name) - client-side filtering
- Chronological order (newest first)
- Notes: free-form text with character limit, optional field

**Tasks:**
- [ ] US-8.1-B: Create operation entry on cipher create
- [ ] US-8.2-B: Create operation entry on cipher update
- [ ] US-8.3-B: Create operation entry on cipher delete
- [ ] US-8.4-B: GET /api/operations endpoint (return all operations with filtering)
- [ ] US-8.5-B: Notes field validation (character limit)
- [ ] US-8.6-B: Integration tests for operations endpoint
- [ ] US-8.1-F: Global audit log page component
- [ ] US-8.2-F: Operations list/timeline component
- [ ] US-8.3-F: Filter input component (by service name)
- [ ] US-8.4-F: Frontend service to fetch operations
- [ ] US-8.5-F: Notes input field (optional, with char limit display)
- [ ] US-8.6-F: Empty state component
- [ ] US-8.7-F: Client-side filtering logic

**Priority:** 9 (Phase 2)

---

### US-9a: React Router Setup
**As a** developer  
**I want to** set up React Router with basic routing  
**So that** we can navigate between pages

**Acceptance Criteria:**
- React Router installed and configured
- Route definitions for all pages
- Basic navigation working

**Tasks:**
- [ ] US-9a.1-F: Install React Router
- [ ] US-9a.2-F: Create router configuration
- [ ] US-9a.3-F: Define routes (login, register, passwords, new, log, user)
- [ ] US-9a.4-F: Basic navigation between routes

**Priority:** 10 (Phase 3)

---

### US-9b: Main Layout Component
**As a** developer  
**I want to** create the main app layout  
**So that** users have a consistent navigation experience

**Acceptance Criteria:**
- Header with logo/brand and user menu
- Top tab navigation with 4 sections
- Content area for page components
- Responsive layout (desktop + mobile)

**Tasks:**
- [ ] US-9b.1-F: Header component (logo + user menu)
- [ ] US-9b.2-F: Top navigation tabs component
- [ ] US-9b.3-F: Main content area wrapper
- [ ] US-9b.4-F: CSS reset and base styles
- [ ] US-9b.5-F: Responsive breakpoints (mobile, tablet, desktop)

**Priority:** 11 (Phase 3)

---

### US-9c: Auth Guard
**As a** developer  
**I want to** protect routes that require authentication  
**So that** unauthenticated users are redirected to login

**Acceptance Criteria:**
- Auth guard wrapper component
- Protected routes redirect to login
- Public routes accessible without auth

**Tasks:**
- [ ] US-9c.1-F: Auth guard component
- [ ] US-9c.2-F: Wrap protected routes with auth guard
- [ ] US-9c.3-F: Redirect logic to login page

**Priority:** 12 (Phase 3)

---

### US-9d: Common Components
**As a** developer  
**I want to** create reusable UI components  
**So that** we have consistent UX across the app

**Acceptance Criteria:**
- Loading/skeleton components
- Toast/notification component
- Empty state component
- Form input components

**Tasks:**
- [ ] US-9d.1-F: Loading/skeleton component
- [ ] US-9d.2-F: Toast/notification component
- [ ] US-9d.3-F: Empty state component
- [ ] US-9d.4-F: Reusable form input components

**Priority:** 13 (Phase 3)

---

## Notes

### Encryption Model
- Two-key system: Login password + Master key
- Login password derives encryption key (Argon2id)
- Master key (random 32-byte) encrypts service passwords (AES-256-GCM)
- Master key encrypted with derived key and stored
- Monolith architecture (no separate microservice)

### Security Principles
- Zero-knowledge architecture (client-side encryption)
- Password re-entry for encryption/decryption operations (MVP)
- Generic error messages for failed login attempts
- HTTPS required for all authenticated requests

### Nice-to-Have Features (Post-MVP)
- Rate limiting on auth endpoints
- CAPTCHA verification
- Email verification
- Account lockout after N failed attempts
- User-configurable session timeout
- Token invalidation/blacklist
- Password strength indicator
- Audit log for password changes
- Email notifications
- Device info in audit log
- Additional encryption algorithms (ChaCha20-Poly1305, etc.)
- Password generator
- Categories/tags for passwords
- Export/import functionality
- Two-factor authentication
