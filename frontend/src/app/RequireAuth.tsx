import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { getRole, getToken, homeByRole } from '../lib/auth'
import type { Role } from '../types'

type RequireAuthProps = {
  roles: Role[]
}

export function RequireAuth({ roles }: RequireAuthProps) {
  const role = getRole()
  const token = getToken()
  const location = useLocation()

  // Временный просмотр кабинета ученика без backend.
  // Работает только в режиме разработки и только с ?demo=1.
  const isStudentDemo =
    import.meta.env.DEV &&
    location.pathname.startsWith('/student') &&
    new URLSearchParams(location.search).get('demo') === '1'

  if (isStudentDemo && roles.includes('student')) {
    return <Outlet />
  }

  if (!token || !role) return <Navigate to="/login" replace />
  if (!roles.includes(role)) {
    return <Navigate to={homeByRole[role]} replace />
  }

  return <Outlet />
}
