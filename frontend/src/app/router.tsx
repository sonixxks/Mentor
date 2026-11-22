import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { AuthLayout } from '../layouts/AuthLayout'
import { LoginPage } from '../pages/auth/LoginPage'
import { RegisterPage } from '../pages/auth/RegisterPage'
import { VerifyEmailPage } from '../pages/auth/VerifyEmailPage'
import { ResetPasswordPage } from '../pages/auth/ResetPasswordPage'
import { OnboardingPage } from '../pages/onboarding/OnboardingPage'
import { CabinetPage } from '../pages/CabinetPage'
import { NotFoundPage } from '../pages/NotFoundPage'
import { RequireAuth } from './RequireAuth'
import { getRole, getToken, homeByRole } from '../lib/auth'

function RequireGuest() {
  const token = getToken()
  const role = getRole()

  if (token && role) {
    return <Navigate to={homeByRole[role]} replace />
  }

  return <Outlet />
}

function RootRedirect() {
  const token = getToken()
  const role = getRole()

  if (token && role) {
    return <Navigate to={homeByRole[role]} replace />
  }

  return <Navigate to="/login" replace />
}

export const router = createBrowserRouter([
  { path: '/', element: <RootRedirect /> },

  {
    element: <RequireGuest />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/register', element: <RegisterPage /> },
    ],
  },

  {
    element: <AuthLayout />,
    children: [
      { path: '/verify-email', element: <VerifyEmailPage /> },
      { path: '/reset-password', element: <ResetPasswordPage /> },
    ],
  },

  {
    element: <RequireAuth roles={['student', 'mentor']} />,
    children: [{ path: '/onboarding/:role/:step', element: <OnboardingPage /> }],
  },

  {
    element: <RequireAuth roles={['student']} />,
    children: [{ path: '/student', element: <CabinetPage /> }],
  },
  {
    element: <RequireAuth roles={['mentor']} />,
    children: [{ path: '/mentor', element: <CabinetPage /> }],
  },
  {
    element: <RequireAuth roles={['curator', 'admin']} />,
    children: [{ path: '/curator', element: <CabinetPage /> }],
  },

  { path: '*', element: <NotFoundPage /> },
])
