import { Outlet, useLocation } from 'react-router-dom'
import { useMyProfile } from '../../hooks/useMyProfile'
import { useSession } from '../../hooks/useSession'
import { CuratorHelpCard } from '../../widgets/CuratorHelpCard'
import { NAV, findActive } from '../../widgets/nav'
import { Sidebar } from '../../widgets/Sidebar'
import { TabBar } from '../../widgets/TabBar'
import { TopBar } from '../../widgets/TopBar'
import styles from './CabinetLayout.module.css'

export function CabinetLayout() {
  const { role, logout } = useSession()
  const { pathname } = useLocation()
  const profile = useMyProfile()

  if (!role) return null

  const items = NAV[role]
  const active = findActive(items, pathname)
  const title = active?.title ?? active?.label ?? ''
  const fullName = [profile.data?.first_name, profile.data?.last_name].filter(Boolean).join(' ')

  return (
    <div className={styles.layout}>
      <Sidebar
        items={items}
        onLogout={logout}
        extra={role === 'student' ? <CuratorHelpCard /> : null}
      />
      <div className={styles.content}>
        <TopBar title={title} fullName={fullName} />
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
      <TabBar items={items} />
    </div>
  )
}
