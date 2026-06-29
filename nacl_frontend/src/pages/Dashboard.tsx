import { Layout } from '../components/Layout'

export const Dashboard = () => {
  return (
    <Layout>
      <div className='dash-stack'>
        <div className='demo-banner'>
          Store real credentials here at your own risk.
        </div>

        <div className='card'>
          <h3 className='card-title'>What is NaCl?</h3>
          <p className='card-text'>
            NaCl is a password manager. You create an account and you store the
            credentials you use to log into other services; NaCl keeps them
            encrypted so they&apos;re unreadable to anyone who gets hold of the
            database. The name comes from the chemical formula for salt, and a
            cryptographic salt is exactly what NaCl uses to protect the key that
            encrypts your data.
          </p>
        </div>

        <div className='card'>
          <h3 className='card-title'>How it works at a glance</h3>
          <p className='card-text'>
            On registration, NaCl generates a random master key and encrypts it
            with a key derived from your password using Argon2id.
            <br />
            When you save a credential, the master key encrypts the service
            username and password with AES-256-GCM.
            <br />
            When you change your login password, only the master key is
            re-encrypted; your stored credentials stay exactly as they are.
          </p>
        </div>

        <div className='card'>
          <h3 className='card-title'>Why encrypt your passwords?</h3>
          <p className='card-text'>
            Storing passwords in your browser&apos;s autosave or with a
            third-party service puts your credentials in someone else&apos;s
            hands. NaCl encrypts each credential with AES-256-GCM before it
            reaches the database, so even if the database leaks, your plaintext
            passwords don&apos;t. NaCl is server-side encrypted: the master key
            exists in backend memory for the duration of each operation.
          </p>
        </div>

        <div className='card'>
          <h3 className='card-title'>The crypto: AES-256-GCM and Argon2id</h3>
          <p className='card-text'>
            <strong>AES-256-GCM</strong>: the symmetric cipher used to encrypt
            your credentials. Authenticated encryption: detects tampering with
            ciphertext on decrypt.
            <br />
            <strong>Argon2id</strong>: the key derivation function. Memory-hard,
            so it slows down GPU brute force. Used to derive the wrapping key
            from your login password and to verify your password at login.
            <br />
            <strong>Envelope encryption</strong>: a two-key design. A random
            master key encrypts your data; your password-derived key only wraps
            the master key, so rotating your password is O(1).
          </p>
        </div>
      </div>
    </Layout>
  )
}
