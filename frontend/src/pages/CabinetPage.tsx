import { useLocation } from 'react-router-dom'
import { Button } from '../components/ui/Button'
import { useSession } from '../hooks/useSession'
import { roleLabel } from '../lib/auth'
import { StudentCabinet } from './student/StudentCabinet'

export function CabinetPage() {
  const { role, logout } = useSession()
  const location = useLocation()

  const isStudentDemo =
    import.meta.env.DEV &&
    location.pathname.startsWith('/student') &&
    new URLSearchParams(location.search).get('demo') === '1'

  if (role === 'student' || isStudentDemo) {
    return <StudentCabinet />
  }

  return (
    <main
      style={{
        padding: 40,
        display: 'flex',
        flexDirection: 'column',
        gap: 16,
        alignItems: 'flex-start',
      }}
    >
      <h1 className="t-h1">Кабинет: {role ? roleLabel[role] : ''}</h1>
      <Button variant="secondary" onClick={logout}>
        Выйти
      </Button>
    </main>
  )
}
