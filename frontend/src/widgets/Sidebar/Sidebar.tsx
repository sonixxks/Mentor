import { clsx } from 'clsx'
import { LogOut } from 'lucide-react'
import type { ReactNode } from 'react'
import { NavLink } from 'react-router-dom'
import { Logo } from '../../components/Logo'
import { isHome } from '../nav'
import type { NavItem } from '../nav'
import styles from './Sidebar.module.css'

type SidebarProps = {
  items: NavItem[]
  onLogout: () => void
  /** Блок над кнопкой «Выйти», например карточка «Не можешь выбрать?» */
  extra?: ReactNode
}

export function Sidebar({ items, onLogout, extra }: SidebarProps) {
  return (
    <nav className={styles.sidebar} aria-label="Основное меню">
      <div className={styles.logo}>
        <Logo />
      </div>

      {items.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          end={isHome(item)}
          className={({ isActive }) => clsx(styles.link, isActive ? styles.active : '')}
        >
          <item.icon size={20} strokeWidth={1.75} className={styles.icon} />
          {item.label}
        </NavLink>
      ))}

      <div className={styles.bottom}>
        {extra}
        <button type="button" className={styles.link} onClick={onLogout}>
          <LogOut size={20} strokeWidth={1.75} />
          Выйти
        </button>
      </div>
    </nav>
  )
}