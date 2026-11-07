import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { CodeInput } from '../../components/ui/CodeInput'
import { useCountdown } from '../../hooks/useCountdown'
import styles from './Auth.module.css'

export function VerifyEmailPage() {
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const role = params.get('role') === 'mentor' ? 'mentor' : 'student'
  const email = params.get('email') ?? 'вашу почту'

  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const timer = useCountdown(60)

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (code.length < 6) {
      setError('Введите все 6 цифр')
      return
    }
    // TODO: POST /api/auth/verify-email — подключим с мок-сервером
    navigate(`/onboarding/${role}/1`)
  }

  function resend() {
    // TODO: POST /api/auth/verify-email/resend
    setCode('')
    setError('')
    timer.restart()
  }

  return (
    <Card variant="glass" className={styles.card}>
      <div className={styles.head}>
        <h1 className="t-h1">Подтвердите почту</h1>
        <p className="t-body text-secondary">
          Отправили код из 6 цифр на <b>{email}</b>. <Link to="/register">Изменить почту</Link>
        </p>
      </div>

      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <CodeInput value={code} onChange={setCode} error={error} />
        <Button type="submit" size="lg" fullWidth>
          Подтвердить и перейти к анкете
        </Button>
      </form>

      <p className={styles.resend}>
        Письмо не пришло? Проверьте папку «Спам».{' '}
        {timer.left > 0 ? (
          <>Запросить новый код можно через {timer.label}</>
        ) : (
          <button type="button" className={styles.linkButton} onClick={resend}>
            Запросить новый код
          </button>
        )}
      </p>
    </Card>
  )
}
