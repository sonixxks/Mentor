import { clsx } from 'clsx'
import { NavLink } from 'react-router-dom'
import { isHome } from '../nav'
import type { NavItem } from '../nav'
import styles from './TabBar.module.css'

export function TabBar({ items }: { items: NavItem[] }) {
  return (
    <nav className={styles.tabbar} aria-label="Нижнее меню">
      {items
        .filter((item) => item.tab)
        .map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={isHome(item)}
            className={({ isActive }) => clsx(styles.tab, isActive ? styles.active : '')}
          >
            <span className={styles.icon}>
              <item.icon size={20} strokeWidth={1.75} />
            </span>
            {item.short ?? item.label}
          </NavLink>
        ))}
    </nav>
  )
}