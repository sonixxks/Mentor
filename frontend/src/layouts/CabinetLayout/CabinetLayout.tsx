import { Outlet, useLocation, useMatches, useNavigate } from 'react-router-dom'
import { useMyProfile } from '../../hooks/useMyProfile'
import { useSession } from '../../hooks/useSession'
import { homeByRole } from '../../lib/auth'
import { greeting } from '../../lib/dates'
import { CuratorHelpCard } from '../../widgets/CuratorHelpCard'
import { NAV, findActive } from '../../widgets/nav'
import { Sidebar } from '../../widgets/Sidebar'
import { TabBar } from '../../widgets/TabBar'
import { TopBar } from '../../widgets/TopBar'
import styles from './CabinetLayout.module.css'

type RouteHandle = { title?: string; back?: boolean; greeting?: boolean }

export function CabinetLayout() {
  const { role, logout } = useSession()
  const location = useLocation()
  const navigate = useNavigate()
  const matches = useMatches()
  const profile = useMyProfile()

  if (!role) return null

  const items = NAV[role]
  const active = findActive(items, location.pathname)
  const handle = matches[matches.length - 1]?.handle as RouteHandle | undefined
  const firstName = profile.data?.first_name

  const title =
    handle?.greeting && firstName
      ? `${greeting()}, ${firstName}`
      : (handle?.title ?? active?.title ?? active?.label ?? '')
  const fullName = [profile.data?.first_name, profile.data?.last_name].filter(Boolean).join(' ')

  function goBack() {
    if (location.key === 'default') navigate(homeByRole[role!])
    else navigate(-1)
  }

  return (
    <div className={styles.layout}>
      <Sidebar
        items={items}
        onLogout={logout}
        extra={role === 'student' ? <CuratorHelpCard /> : null}
      />
      <div className={styles.content}>
        <TopBar title={title} fullName={fullName} onBack={handle?.back ? goBack : undefined} />
        <main className={styles.main}>
          <Outlet />
        </main>
      </div>
      <TabBar items={items} />
    </div>
  )
}
