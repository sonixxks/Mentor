import { Outlet, useLocation } from 'react-router-dom'
import { useSession } from '../../hooks/useSession'
import { CuratorHelpCard } from '../../widgets/CuratorHelpCard'
import { NAV, findActive } from '../../widgets/nav'
import { Sidebar } from '../../widgets/Sidebar'
import { TabBar } from '../../widgets/TabBar'
import { TopBar } from '../../widgets/TopBar'
import styles from './CabinetLayout.module.css'

// TODO: имя из GET /profiles/me, когда подключим MSW
const FULL_NAME = 'Екатерина Смирнова'

export function CabinetLayout() {
  const { role, logout } = useSession()
  const { pathname } = useLocation()

  if (!role) return null

  const items = NAV[role]
  const active = findActive(items, pathname)
  const title = active?.title ?? active?.label ?? ''

  return (
    <div className={styles.layout}>
      <Sidebar
        items={items}
        onLogout={logout}
        extra={role === 'student' ? <CuratorHelpCard /> : null}
      />
      <div className={styles.content}>
        <TopBar title={title} fullName={FULL_NAME} />
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
      <TabBar items={items} />
    </div>
  )
}
