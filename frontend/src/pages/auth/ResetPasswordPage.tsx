import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { CodeInput } from '../../components/ui/CodeInput'
import { Input } from '../../components/ui/Input'
import { useCountdown } from '../../hooks/useCountdown'
import { isEmail, PASSWORD_MIN } from '../../lib/validation'
import {
  requestPasswordReset,
  verifyPasswordResetCode,
  resetPassword,
  getMyAccount,
} from '../../api/auth'
import { ApiError } from '../../api/client'
import { clearSession, saveSession, updateTokens } from '../../lib/auth'
import styles from './Auth.module.css'
import { pathAfterLogin } from '../../lib/afterLogin'

type Step = 'email' | 'code' | 'password'
type Errors = { email?: string; code?: string; password?: string; repeat?: string }

const STEP_NUMBER: Record<Step, number> = { email: 1, code: 2, password: 3 }

const TITLES: Record<Step, string> = {
  email: 'Забыли пароль?',
  code: 'Введите код',
  password: 'Новый пароль',
}

export function ResetPasswordPage() {
  const navigate = useNavigate()

  const [step, setStep] = useState<Step>('email')
  const [email, setEmail] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [repeat, setRepeat] = useState('')
  const [errors, setErrors] = useState<Errors>({})
  const [serverError, setServerError] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isResending, setIsResending] = useState(false)

  const timer = useCountdown(60)

  async function sendCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    setErrors({})
    setServerError('')

    if (!isEmail(email)) {
      setErrors({ email: 'Введите почту в формате name@mail.ru' })
      return
    }

    setIsLoading(true)

    try {
      await requestPasswordReset(email.trim().toLowerCase())

      setEmail(email.trim().toLowerCase())
      setStep('code')
      setCode('')
      timer.restart()
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.status === 429) {
          setServerError('Слишком много запросов. Попробуйте позже.')
        } else {
          setServerError(error.message)
        }
      } else {
        setServerError('Не удалось отправить код. Попробуйте позже.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  async function checkCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    setErrors({})
    setServerError('')

    if (!/^\d{6}$/.test(code)) {
      setErrors({ code: 'Введите все 6 цифр' })
      return
    }

    setIsLoading(true)

    try {
      await verifyPasswordResetCode(email, code)
      setStep('password')
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.status === 400) {
          setErrors({
            code: 'Неверный или просроченный код. Проверьте его.',
          })
        } else if (error.status === 429) {
          setServerError('Слишком много попыток. Попробуйте позже.')
        } else {
          setServerError(error.message)
        }
      } else {
        setServerError('Не удалось проверить код. Попробуйте позже.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  async function resend() {
    if (isResending || isLoading || timer.left > 0) return

    setIsResending(true)
    setErrors({})
    setServerError('')

    try {
      await requestPasswordReset(email)
      setCode('')
      timer.restart()
    } catch (error) {
      if (error instanceof ApiError) {
        if (error.status === 429) {
          setServerError('Повторно запросить код пока нельзя. Попробуйте позже.')
        } else {
          setServerError(error.message)
        }
      } else {
        setServerError('Не удалось отправить код. Попробуйте позже.')
      }
    } finally {
      setIsResending(false)
    }
  }

  async function savePassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const next: Errors = {}

    if (password.length < PASSWORD_MIN) {
      next.password = `Минимум ${PASSWORD_MIN} символов`
    }

    if (repeat !== password) {
      next.repeat = 'Пароли не совпадают'
    }

    setErrors(next)
    setServerError('')

    if (Object.keys(next).length > 0) return

    setIsLoading(true)

    try {
      const tokens = await resetPassword(email, code, password)

      // Сохраняем новые токены перед запросом профиля.
      updateTokens(tokens.access_token, tokens.refresh_token)

      // Получаем роль пользователя с сервера.
      const account = await getMyAccount()

      saveSession(tokens.access_token, tokens.refresh_token, account.role)

      navigate(await pathAfterLogin(account.role), { replace: true })
    } catch (error) {
      clearSession()

      if (error instanceof ApiError) {
        if (error.status === 400) {
          setServerError('Не удалось изменить пароль. Возможно, код истёк.')
        } else if (error.status === 429) {
          setServerError('Слишком много попыток. Попробуйте позже.')
        } else {
          setServerError(error.message)
        }
      } else {
        setServerError('Не удалось сохранить пароль. Попробуйте позже.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Card variant="glass" className={styles.card}>
      <div className={styles.head}>
        <span className={styles.step}>Шаг {STEP_NUMBER[step]} из 3</span>
        <h1 className="t-h1">{TITLES[step]}</h1>

        {step === 'email' ? (
          <p className="t-body text-secondary">
            Укажите почту, с которой регистрировались. Пришлём код из 6 цифр.
          </p>
        ) : null}

        {step === 'code' ? (
          <p className="t-body text-secondary">
            Отправили код на <b>{email}</b>. Он действует 10 минут.
          </p>
        ) : null}

        {step === 'password' ? (
          <p className="t-body text-secondary">Придумайте новый пароль для входа.</p>
        ) : null}
      </div>

      {step === 'email' ? (
        <form className={styles.form} onSubmit={sendCode} noValidate>
          <Input
            label="Электронная почта"
            type="email"
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            error={errors.email}
          />

          {serverError ? <p className={styles.formError}>{serverError}</p> : null}

          <Button type="submit" size="lg" fullWidth disabled={isLoading}>
            {isLoading ? 'Отправляем...' : 'Получить код'}
          </Button>
        </form>
      ) : null}

      {step === 'code' ? (
        <form className={styles.form} onSubmit={checkCode} noValidate>
          <CodeInput value={code} onChange={setCode} error={errors.code} />

          {serverError ? <p className={styles.formError}>{serverError}</p> : null}

          <p className={styles.resend}>
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

          <Button type="submit" size="lg" fullWidth disabled={isLoading || isResending}>
            {isLoading ? 'Проверяем...' : 'Продолжить'}
          </Button>
        </form>
      ) : null}

      {step === 'password' ? (
        <form className={styles.form} onSubmit={savePassword} noValidate>
          <Input
            label="Новый пароль"
            type="password"
            autoComplete="new-password"
            placeholder={`Минимум ${PASSWORD_MIN} символов`}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={errors.password}
          />

          <Input
            label="Повторите пароль"
            type="password"
            autoComplete="new-password"
            value={repeat}
            onChange={(e) => setRepeat(e.target.value)}
            error={errors.repeat}
          />

          {serverError ? <p className={styles.formError}>{serverError}</p> : null}

          <Button type="submit" size="lg" fullWidth disabled={isLoading}>
            {isLoading ? 'Сохраняем...' : 'Сохранить и войти'}
          </Button>
        </form>
      ) : null}

      <Link to="/login" className={styles.footer}>
        Вернуться ко входу
      </Link>
    </Card>
  )
}
