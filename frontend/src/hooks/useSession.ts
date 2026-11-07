import { useNavigate } from 'react-router-dom'
import { clearSession, getRole, getToken } from '../lib/auth'

export function useSession() {
  const navigate = useNavigate()

  function logout() {
    clearSession()
    navigate('/login', { replace: true })
  }

  return { token: getToken(), role: getRole(), logout }
}