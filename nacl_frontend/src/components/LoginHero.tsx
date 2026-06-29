export const LoginHero = () => {
  return (
    <div className='login-hero'>
      <h2>Encrypt and manage your passwords on your own terms</h2>
      <p>
        NaCl is a password manager. Your credentials are encrypted with
        AES-256-GCM before they reach the database, using a master key derived
        from your login password with Argon2id. When you change your password,
        only the master key is re-encrypted; your stored credentials stay
        untouched.
      </p>
    </div>
  )
}
