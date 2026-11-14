import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Card } from '../../components/ui/Card'
import { CodeInput } from '../../components/ui/CodeInput'
import { Input } from '../../components/ui/Input'
import { useCountdown } from '../../hooks/useCountdown'
import { isEmail, PASSWORD_MIN } from '../../lib/validation'
import styles from './Auth.module.css'

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
  const timer = useCountdown(60)

  function sendCode(event: FormEvent) {
    event.preventDefault()
    if (!isEmail(email)) {
      setErrors({ email: 'Введите почту в формате name@mail.ru' })
      return
    }
    // TODO: POST /api/auth/reset-password/request
    setErrors({})
    setStep('code')
    timer.restart()
  }

  function checkCode(event: FormEvent) {
    event.preventDefault()
    if (code.length < 6) {
      setErrors({ code: 'Введите все 6 цифр' })
      return
    }
    // TODO: POST /api/auth/reset-password/verify — сервер скажет, верный ли код
    setErrors({})
    setStep('password')
  }

  function resend() {
    // TODO: POST /api/auth/reset-password/request
    setCode('')
    setErrors({})
    timer.restart()
  }

  function savePassword(event: FormEvent) {
    event.preventDefault()
    const next: Errors = {}
    if (password.length < PASSWORD_MIN) next.password = `Минимум ${PASSWORD_MIN} символов`
    if (repeat !== password) next.repeat = 'Пароли не совпадают'
    setErrors(next)
    if (Object.keys(next).length > 0) return
    // TODO: POST /api/auth/reset-password/confirm
    navigate('/login')
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
          <Button type="submit" size="lg" fullWidth>
            Получить код
          </Button>
        </form>
      ) : null}

      {step === 'code' ? (
        <form className={styles.form} onSubmit={checkCode} noValidate>
          <CodeInput value={code} onChange={setCode} error={errors.code} />
          <p className={styles.resend}>
            {timer.left > 0 ? (
              <>Запросить новый код можно через {timer.label}</>
            ) : (
              <button type="button" className={styles.linkButton} onClick={resend}>
                Запросить новый код
              </button>
            )}
          </p>
          <Button type="submit" size="lg" fullWidth>
            Продолжить
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
          <Button type="submit" size="lg" fullWidth>
            Сохранить и войти
          </Button>
        </form>
      ) : null}

      <Link to="/login" className={styles.footer}>
        Вернуться ко входу
      </Link>
    </Card>
  )
}