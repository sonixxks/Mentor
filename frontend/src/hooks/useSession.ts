import { useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { clearSession, getRole, getToken } from '../lib/auth'

export function useSession() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()

  function logout() {
    clearSession()
    queryClient.clear()
    navigate('/login', { replace: true })
  }

  return { token: getToken(), role: getRole(), logout }
}
