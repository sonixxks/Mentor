import { Bell } from 'lucide-react'
import { Avatar } from '../../components/ui/Avatar'
import styles from './TopBar.module.css'

type TopBarProps = {
  title: string
  fullName: string
}

export function TopBar({ title, fullName }: TopBarProps) {
  const firstName = fullName.split(' ')[0]

  return (
    <header className={styles.topbar}>
      <h1 className={styles.title}>{title}</h1>
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
