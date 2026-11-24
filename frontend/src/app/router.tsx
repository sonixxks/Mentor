import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom'
import { AuthLayout } from '../layouts/AuthLayout'
import { LoginPage } from '../pages/auth/LoginPage'
import { RegisterPage } from '../pages/auth/RegisterPage'
import { VerifyEmailPage } from '../pages/auth/VerifyEmailPage'
import { ResetPasswordPage } from '../pages/auth/ResetPasswordPage'
import { OnboardingPage } from '../pages/onboarding/OnboardingPage'
import { CabinetLayout } from '../layouts/CabinetLayout'
import { ComingSoonPage } from '../pages/ComingSoonPage'
import { StudentCabinet } from '../pages/student/StudentCabinet'
import { StudentCatalog } from '../pages/student/StudentCatalog'
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
    children: [
      {
        element: <CabinetLayout />,
        children: [
          { path: '/student', element: <StudentCabinet /> },
          { path: '/student/catalog', element: <StudentCatalog /> },
          { path: '/student/*', element: <ComingSoonPage /> },
        ],
      },
    ],
  },
  {
    element: <RequireAuth roles={['mentor']} />,
    children: [
      {
        element: <CabinetLayout />,
        children: [
          { path: '/mentor', element: <ComingSoonPage /> },
          { path: '/mentor/*', element: <ComingSoonPage /> },
        ],
      },
    ],
  },
  {
    element: <RequireAuth roles={['curator', 'admin']} />,
    children: [
      {
        element: <CabinetLayout />,
        children: [
          { path: '/curator', element: <ComingSoonPage /> },
          { path: '/curator/*', element: <ComingSoonPage /> },
        ],
      },
    ],
  },

  { path: '*', element: <NotFoundPage /> },
])
