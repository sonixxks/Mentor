import { Button } from '../components/ui/Button'

export function NotFoundPage() {
  return (
    <main style={{ minHeight: '100vh', display: 'grid', placeItems: 'center', alignContent: 'center', gap: 16 }}>
      <span className="t-display">404</span>
      <p className="t-body text-secondary">Такой страницы нет</p>
      <Button to="/">На главную</Button>
    </main>
  )
}