import { Outlet } from 'react-router-dom'
import { Logo } from '../../components/Logo'
import styles from './AuthLayout.module.css'

export function AuthLayout() {
  return (
    <div className={styles.layout}>
      <div className={styles.glow} />
      <header className={styles.header}>
        <Logo to="/" />
      </header>
      <main className={styles.content}>
        <Outlet />
      </main>
    </div>
  )
}
