import { useNavigate } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { homeByRole, roleLabel, saveSession } from '../../lib/auth'
import type { Role } from '../../types'

const DEV_ROLES: Role[] = ['student', 'mentor', 'curator']

export function LoginPage() {
  const navigate = useNavigate()

  function devLogin(role: Role) {
    saveSession('dev-token', role)
    navigate(homeByRole[role])
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h1 className="t-h1">Вход</h1>
      <div style={{ display: 'flex', gap: 8 }}>
        {DEV_ROLES.map((role) => (
          <Button key={role} variant="secondary" size="sm" onClick={() => devLogin(role)}>
            {roleLabel[role]}
          </Button>
        ))}
      </div>
    </div>
  )
}