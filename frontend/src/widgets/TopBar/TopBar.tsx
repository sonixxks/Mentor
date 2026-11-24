import { ArrowLeft, Bell } from 'lucide-react'
import { Avatar } from '../../components/ui/Avatar'
import styles from './TopBar.module.css'

type TopBarProps = {
  title: string
  /** Полное имя: по нему Avatar рисует инициалы */
  fullName: string
  /** Если передан, слева от заголовка появляется стрелка «Назад» */
  onBack?: () => void
}

export function TopBar({ title, fullName, onBack }: TopBarProps) {
  const firstName = fullName.split(' ')[0]

  return (
    <header className={styles.topbar}>
      <div className={styles.titleRow}>
        {onBack ? (
          <button type="button" className={styles.back} onClick={onBack} aria-label="Назад">
            <ArrowLeft size={22} strokeWidth={1.75} />
          </button>
        ) : null}
        <h1 className={styles.title}>{title}</h1>
      </div>
      <div className={styles.user}>
        {/* TODO: уведомления */}
        <button type="button" className={styles.bell} aria-label="Уведомления">
          <Bell size={20} strokeWidth={1.75} />
        </button>
        <span className={styles.name}>{firstName}</span>
        <Avatar name={fullName} size="s" />
      </div>
    </header>
  )
}
