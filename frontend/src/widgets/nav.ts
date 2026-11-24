import {
  Calendar,
  Inbox,
  LayoutGrid,
  MessageCircle,
  Search,
  ShieldCheck,
  Shuffle,
  Sparkles,
  User,
  Users,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'
import type { Role } from '../types'

export type NavItem = {
  to: string
  label: string
  title?: string
  short?: string
  icon: LucideIcon
  tab?: boolean
}

const STUDENT: NavItem[] = [
  { to: '/student', label: 'Подбор', icon: Sparkles, tab: true },
  { to: '/student/catalog', label: 'Каталог', title: 'Каталог наставников', icon: Search, tab: true },
  { to: '/student/requests', label: 'Мои заявки', short: 'Заявки', icon: Inbox, tab: true },
  { to: '/student/workspace', label: 'Пространство', short: 'Задачи', icon: LayoutGrid, tab: true },
  { to: '/student/calendar', label: 'Календарь', icon: Calendar },
  { to: '/student/chat', label: 'Чат', icon: MessageCircle, tab: true },
]

const MENTOR: NavItem[] = [
  { to: '/mentor', label: 'Дашборд', short: 'Главная', icon: LayoutGrid, tab: true },
  { to: '/mentor/requests', label: 'Заявки', title: 'Входящие заявки', icon: Inbox, tab: true },
  { to: '/mentor/students', label: 'Ученики', icon: Users, tab: true },
  { to: '/mentor/calendar', label: 'Календарь', icon: Calendar },
  { to: '/mentor/chat', label: 'Чат', icon: MessageCircle, tab: true },
  { to: '/mentor/profile', label: 'Профиль', icon: User, tab: true },
]

const CURATOR: NavItem[] = [
  { to: '/curator', label: 'Пары', title: 'Светофор пар', icon: Users, tab: true },
  { to: '/curator/matching', label: 'Ручной подбор', short: 'Подбор', icon: Shuffle, tab: true },
  { to: '/curator/verify', label: 'Проверка опыта', short: 'Проверка', icon: ShieldCheck, tab: true },
]

export const NAV: Record<Role, NavItem[]> = {
  student: STUDENT,
  mentor: MENTOR,
  curator: CURATOR,
  admin: CURATOR,
}

export function isHome(item: NavItem) {
  return item.to.split('/').length === 2
}

export function findActive(items: NavItem[], pathname: string) {
  return items
    .filter((item) => pathname === item.to || pathname.startsWith(item.to + '/'))
    .sort((a, b) => b.to.length - a.to.length)[0]
}