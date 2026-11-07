import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Check, User, Users } from 'lucide-react'
import { clsx } from 'clsx'
import { Button } from '../../components/ui/Button'
import { AuthSplit } from '../../layouts/AuthSplit'
import { Input } from '../../components/ui/Input'
import { isEmail, PASSWORD_MIN } from '../../lib/validation'
import type { PublicRole } from '../../types'
import styles from './Auth.module.css'

type Field = 'lastName' | 'firstName' | 'email' | 'password' | 'repeat' | 'agree'
type Errors = Partial<Record<Field, string>>

const ROLE_OPTIONS = [
  { role: 'student', title: 'Ученик', text: 'Ищу наставника, чтобы расти в профессии', icon: User },
  {
    role: 'mentor',
    title: 'Наставник',
    text: 'Готов делиться опытом и вести учеников',
    icon: Users,
  },
] as const

const PROMO: Record<PublicRole, { title: string; text: string }> = {
  student: {
    title: 'Найди своего наставника',
    text: 'Пять минут на анкету — и алгоритм подберёт наставников под твои цели, навыки и свободное время.',
  },
  mentor: {
    title: 'Делитесь опытом',
    text: 'Расскажите о себе, и платформа предложит учеников, которым ваш опыт подойдёт лучше всего.',
  },
}

export function RegisterPage() {
  const navigate = useNavigate()
  const [role, setRole] = useState<PublicRole>('student')
  const [lastName, setLastName] = useState('')
  const [firstName, setFirstName] = useState('')
  const [middleName, setMiddleName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [repeat, setRepeat] = useState('')
  const [agree, setAgree] = useState(false)
  const [errors, setErrors] = useState<Errors>({})

  function validate(): Errors {
    const next: Errors = {}
    if (!lastName.trim()) next.lastName = 'Укажите фамилию'
    if (!firstName.trim()) next.firstName = 'Укажите имя'
    if (!isEmail(email)) next.email = 'Введите почту в формате name@mail.ru'
    if (password.length < PASSWORD_MIN) next.password = `Минимум ${PASSWORD_MIN} символов`
    if (repeat !== password) next.repeat = 'Пароли не совпадают'
    if (!agree) next.agree = 'Нужно согласие с условиями'
    return next
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const next = validate()
    setErrors(next)
    if (Object.keys(next).length > 0) return
    // TODO: POST /api/auth/register — подключим с мок-сервером
    navigate(`/verify-email?role=${role}&email=${encodeURIComponent(email.trim())}`)
  }

  return (
    <AuthSplit title={PROMO[role].title} text={PROMO[role].text}>
      <div className={styles.head}>
        <h1 className="t-h1">Регистрация</h1>
        <p className={styles.footer}>
          Уже есть аккаунт? <Link to="/login">Войти</Link>
        </p>
      </div>

      <div className={styles.roles}>
        {ROLE_OPTIONS.map(({ role: value, title, text, icon: Icon }) => (
          <button
            key={value}
            type="button"
            className={clsx(styles.role, role === value ? styles.roleActive : '')}
            onClick={() => setRole(value)}
          >
            <span className={styles.roleIcon}>
              {role === value ? <Check size={20} /> : <Icon size={20} strokeWidth={1.75} />}
            </span>
            <span className={styles.roleTitle}>{title}</span>
            <span className={styles.roleText}>{text}</span>
          </button>
        ))}
      </div>

      <form className={styles.form} onSubmit={handleSubmit} noValidate>
        <div className={styles.grid3}>
          <Input
            label="Фамилия"
            autoComplete="family-name"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            error={errors.lastName}
          />
          <Input
            label="Имя"
            autoComplete="given-name"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            error={errors.firstName}
          />
          <Input
            label="Отчество"
            autoComplete="additional-name"
            placeholder="Если есть"
            value={middleName}
            onChange={(e) => setMiddleName(e.target.value)}
          />
        </div>
        <Input
          label="Электронная почта"
          type="email"
          autoComplete="email"
          placeholder="ekaterina@urfu.me"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={errors.email}
        />
        <div className={styles.grid2}>
          <Input
            label="Пароль"
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
        </div>

        <label className={styles.check}>
          <input type="checkbox" checked={agree} onChange={(e) => setAgree(e.target.checked)} />
          <span>
            Согласен(на) с условиями и обработкой персональных данных. Наставники видят только имя и
            интересы ученика.
          </span>
        </label>
        {errors.agree ? <p className={styles.formError}>{errors.agree}</p> : null}

        <Button type="submit" size="lg" fullWidth>
          Создать аккаунт
        </Button>
      </form>
    </AuthSplit>
  )
}
