import type { ReactNode } from 'react'
import { Logo } from '../../components/Logo'
import styles from './AuthSplit.module.css'

type AuthSplitProps = {
  title: string
  text?: string
  extra?: ReactNode
  children: ReactNode
}

export function AuthSplit({ title, text, extra, children }: AuthSplitProps) {
  return (
    <div className={styles.page}>
      <aside className={styles.aside}>
        <div className={styles.glow} aria-hidden="true" />
        <Logo to="/" />
        <div className={styles.intro}>
          <h2 className={styles.title}>{title}</h2>
          {text ? <p className={styles.text}>{text}</p> : null}
        </div>
        {extra ? <div className={styles.extra}>{extra}</div> : null}
      </aside>

      <main className={styles.main}>
        <div className={styles.mobileLogo}>
          <Logo to="/" />
        </div>
        <div className={styles.form}>{children}</div>
      </main>
    </div>
  )
}