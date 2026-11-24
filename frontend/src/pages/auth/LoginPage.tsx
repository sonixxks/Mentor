import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { AuthSplit } from '../../layouts/AuthSplit'
import { isEmail } from '../../lib/validation'
import { clearSession, saveSession, updateTokens } from '../../lib/auth'
import { pathAfterLogin } from '../../lib/afterLogin'
import { login, getMyAccount } from '../../api/auth'
import { ApiError } from '../../api/client'
import styles from './Auth.module.css'

type Errors = { email?: string; password?: string; form?: string }

export function LoginPage() {
  const navigate = useNavigate()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [remember, setRemember] = useState(true)
  const [errors, setErrors] = useState<Errors>({})
  const [isLoading, setIsLoading] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const next: Errors = {}

    if (!isEmail(email)) {
      next.email = 'Введите почту в формате name@mail.ru'
    }

    if (!password) {
      next.password = 'Введите пароль'
    }

    if (Object.keys(next).length > 0) {
      setErrors(next)
      return
    }

    setErrors({})
    setIsLoading(true)

    try {
      const tokens = await login({
        email: email.trim().toLowerCase(),
        password,
        remember_me: remember,
      })

      // Сначала сохраняем токены, чтобы следующий запрос прошёл с авторизацией.
      updateTokens(tokens.access_token, tokens.refresh_token)

      // Получаем роль пользователя с сервера.
      const account = await getMyAccount()

      // Сохраняем сессию вместе с подтверждённой ролью.
      saveSession(tokens.access_token, tokens.refresh_token, account.role)

      navigate(await pathAfterLogin(account.role), { replace: true })
    } catch (error) {
      clearSession()

      if (error instanceof ApiError) {
        if (error.status === 401) {
          setErrors({
            form: 'Неверная почта или пароль. Проверьте введённые данные.',
          })
        } else if (error.status === 429) {
          setErrors({
            form: 'Слишком много попыток входа. Попробуйте позже.',
          })
        } else {
          setErrors({ form: error.message })
        }
      } else {
        setErrors({
          form: 'Не удалось связаться с сервером. Попробуйте позже.',
        })
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <AuthSplit
      title="С возвращением"
      text="Твой наставник, задачи и прогресс ждут в рабочем пространстве."
    >
      <div className={styles.head}>
        <h1 className="t-h1">Вход</h1>
        <p className={styles.footer}>
          Нет аккаунта? <Link to="/register">Зарегистрироваться</Link>
        </p>
      </div>

      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <Input
          label="Электронная почта"
          type="email"
          autoComplete="email"
          placeholder="ekaterina@urfu.me"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={errors.email}
        />

        <Input
          label="Пароль"
          labelAction={<Link to="/reset-password">Забыли пароль?</Link>}
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={errors.password}
        />

        <label className={styles.check}>
          <input
            type="checkbox"
            checked={remember}
            onChange={(e) => setRemember(e.target.checked)}
            disabled={isLoading}
          />
          <span>Запомнить меня</span>
        </label>

        {errors.form ? (
          <p className={styles.formError}>{errors.form}</p>
        ) : null}

        <Button type="submit" size="lg" fullWidth disabled={isLoading}>
          {isLoading ? 'Входим...' : 'Войти'}
        </Button>
      </form>

      {/* TODO: вход через Яндекс ID (POST /auth/yandex) */}

      <p className={styles.note}>
        Аккаунты кураторов создаёт администратор. Если вы куратор, используйте выданные данные для
        входа.
      </p>
    </AuthSplit>
  )
}
