import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { AuthSplit } from '../../layouts/AuthSplit'
import { isEmail } from '../../lib/validation'
import styles from './Auth.module.css'

type Errors = { email?: string; password?: string; form?: string }

export function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [remember, setRemember] = useState(true)
  const [errors, setErrors] = useState<Errors>({})

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const next: Errors = {}
    if (!isEmail(email)) next.email = 'Введите почту в формате name@mail.ru'
    if (!password) next.password = 'Введите пароль'
    if (Object.keys(next).length === 0) {
      // TODO: POST /api/auth/login { email, password, remember } — подключим с мок-сервером
      next.form = 'Вход заработает после подключения сервера.'
    }
    setErrors(next)
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
          />
          <span>Запомнить меня</span>
        </label>
        {errors.form ? <p className={styles.formError}>{errors.form}</p> : null}
        <Button type="submit" size="lg" fullWidth>
          Войти
        </Button>
      </form>

      <span className={styles.divider}>или</span>
      {/* TODO: вход через Яндекс ID */}
      <Button variant="secondary" size="lg" fullWidth>
        Войти через Яндекс ID
      </Button>

      <p className={styles.note}>
        Аккаунты кураторов создаёт администратор. Если вы куратор, используйте выданные данные для
        входа.
      </p>
    </AuthSplit>
  )
}
