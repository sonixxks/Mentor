import { Navigate, Outlet } from 'react-router-dom'
import { getRole, getToken, homeByRole } from '../lib/auth'
import type { Role } from '../types'

type RequireAuthProps = {
  roles: Role[]
}

export function RequireAuth({ roles }: RequireAuthProps) {
  const role = getRole()

  if (!getToken() || !role) return <Navigate to="/login" replace />
  if (!roles.includes(role)) return <Navigate to={homeByRole[role]} replace />

  return <Outlet />
}