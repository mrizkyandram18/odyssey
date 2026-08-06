import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { HashRouter } from 'react-router-dom'
import './index.css'
import { App } from './app/App'
import { SessionProvider } from './app/SessionProvider'
import { HttpAuthClient } from './shared/lib/auth'
import { apiClient } from './shared/lib/api'
import { registerServiceWorker } from './app/service-worker-registration'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <SessionProvider authClient={new HttpAuthClient(apiClient)}>
      <HashRouter>
        <App />
      </HashRouter>
    </SessionProvider>
  </StrictMode>,
)

registerServiceWorker()
