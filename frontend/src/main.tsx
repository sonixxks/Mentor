import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './styles/tokens.css'
import './styles/global.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <main>
      <h1 className="t-h1">Куратор</h1>
      <p className="t-body text-secondary">В разработке</p>
    </main>
  </StrictMode>,
)