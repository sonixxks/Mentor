import { Navigate, useParams } from 'react-router-dom'
import { MentorOnboarding } from './mentor/MentorOnboarding'
import { StudentOnboarding } from './student/StudentOnboarding'
import { STEPS } from './steps'

export function OnboardingPage() {
  const { role, step } = useParams()

  if (role !== 'student' && role !== 'mentor') {
    return <Navigate to="/register" replace />
  }

  const number = Number(step)
  if (!Number.isInteger(number) || number < 1 || number > STEPS[role].length) {
    return <Navigate to={`/onboarding/${role}/1`} replace />
  }

  return role === 'student' ? (
    <StudentOnboarding step={number} />
  ) : (
    <MentorOnboarding step={number} />
  )
}