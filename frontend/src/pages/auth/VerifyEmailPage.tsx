import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { CodeInput } from '../../components/ui/CodeInput'
import { useCountdown } from '../../hooks/useCountdown'
import { verifyEmail, resendEmailCode, getMyAccount } from '../../api/auth'
import { ApiError } from '../../api/client'
import { saveSession, updateTokens, clearSession, homeByRole } from '../../lib/auth'
import styles from './Auth.module.css'

export function VerifyEmailPage() {
  const navigate = useNavigate()
  const [params] = useSearchParams()

  const email = params.get('email') ?? ''

  const [code, setCode] = useState('')
  const [error, setError] = useState('')
  const [serverError, setServerError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isResending, setIsResending] = useState(false)

  const timer = useCountdown(60)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    setError('')
    setServerError('')

    if (!email) {
      setServerError('Не указана почта. Вернитесь к регистрации.')
      return
    }

    if (!/^\d{6}$/.test(code)) {
      setError('Введите все 6 цифр')
      return
    }

    setIsLoading(true)

    try {
      const tokens = await verifyEmail(email, code)

      // Сохраняем токены, чтобы запрос профиля прошёл с авторизацией.
      updateTokens(tokens.access_token, tokens.refresh_token)

      // Получаем настоящую роль пользователя с сервера.
      const account = await getMyAccount()

      // Сохраняем сессию с ролью, полученной от API.
      saveSession(tokens.access_token, tokens.refresh_token, account.role)

      // После регистрации ученик и наставник переходят к своей анкете.
      if (account.role === 'student' || account.role === 'mentor') {
        navigate(`/onboarding/${account.role}/1`, { replace: true })
      } else {
        navigate(homeByRole[account.role], { replace: true })
      }
    } catch (err) {
      clearSession()

      if (err instanceof ApiError) {
        if (err.status === 400) {
          setError('Неверный или просроченный код. Проверьте его и попробуйте ещё раз.')
        } else if (err.status === 429) {
          setServerError('Слишком много попыток. Попробуйте позже.')
        } else {
          setServerError(err.message)
        }
      } else {
        setServerError('Не удалось связаться с сервером. Попробуйте позже.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  async function resend() {
    if (!email || isResending || isLoading || timer.left > 0) return

    setIsResending(true)
    setError('')
    setServerError('')

    try {
      await resendEmailCode(email)
      setCode('')
      timer.restart()
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 429) {
          setServerError('Повторно запросить код пока нельзя. Попробуйте позже.')
        } else {
          setServerError(err.message)
        }
      } else {
        setServerError('Не удалось отправить код. Попробуйте позже.')
      }
    } finally {
      setIsResending(false)
    }
  }

  return (
    <Card variant="glass" className={styles.card}>
      <div className={styles.head}>
        <h1 className="t-h1">Подтвердите почту</h1>
        <p className="t-body text-secondary">
          Отправили код из 6 цифр на <b>{email || 'вашу почту'}</b>.{' '}
          <Link to="/register">Изменить почту</Link>
        </p>
      </div>

      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <CodeInput value={code} onChange={setCode} error={error} />

        {serverError ? <p className={styles.formError}>{serverError}</p> : null}

        <Button type="submit" size="lg" fullWidth disabled={isLoading || isResending}>
          {isLoading ? 'Проверяем код...' : 'Подтвердить и перейти к анкете'}
        </Button>
      </form>

      <p className={styles.resend}>
        Письмо не пришло? Проверьте папку «Спам».{' '}
        {timer.left > 0 ? (
          <>Запросить новый код можно через {timer.label}</>
        ) : (
          <button
            type="button"
            className={styles.linkButton}
            onClick={resend}
            disabled={isResending || isLoading}
          >
            {isResending ? 'Отправляем...' : 'Запросить новый код'}
          </button>
        )}
      </p>
    </Card>
  )
}
