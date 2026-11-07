import { Button } from '../components/ui/Button'
import { useSession } from '../hooks/useSession'
import { roleLabel } from '../lib/auth'

export function CabinetPage() {
  const { role, logout } = useSession()

  return (
    <main style={{ padding: 40, display: 'flex', flexDirection: 'column', gap: 16, alignItems: 'flex-start' }}>
      <h1 className="t-h1">Кабинет: {role ? roleLabel[role] : ''}</h1>
      <Button variant="secondary" onClick={logout}>
        Выйти
      </Button>
    </main>
  )
}