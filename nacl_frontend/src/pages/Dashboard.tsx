import { Layout } from '../components/Layout'

export const Dashboard = () => {
  return (
    <Layout>
      <div className='card'>
        <h3 className='card-title'>Why encrypt your own passwords?</h3>
        <p className='card-text'>
          Storing passwords in your browser or with third-party services puts
          your credentials in someone else&apos;s hands. NaCl encrypts
          everything on your device before it ever reaches the server, so only
          you hold the key.
        </p>
      </div>
    </Layout>
  )
}
