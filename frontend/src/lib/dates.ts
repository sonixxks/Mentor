const TZ = 'Asia/Yekaterinburg'
const OFFSET = 5 * 60 * 60 * 1000
const DAY = 24 * 60 * 60 * 1000

export function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    timeZone: TZ,
  })
}

export function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: TZ,
  })
}

export function formatMeetingTime(iso: string) {
  const day = new Date(iso).toLocaleDateString('ru-RU', {
    weekday: 'short',
    day: 'numeric',
    month: 'long',
    timeZone: TZ,
  })
  return `${day[0].toUpperCase()}${day.slice(1)} · ${formatTime(iso)}`
}

export function minutesBetween(start: string, end?: string) {
  return end ? Math.round((Date.parse(end) - Date.parse(start)) / 60000) : null
}

export function isPast(iso: string) {
  return Date.parse(iso) < Date.now()
}

export function ekbDate(iso: string) {
  return new Date(Date.parse(iso) + OFFSET).toISOString().slice(0, 10)
}

export function ekbTime(iso: string) {
  return new Date(Date.parse(iso) + OFFSET).toISOString().slice(11, 16)
}

export function fromEkb(date: string, time: string) {
  return new Date(`${date}T${time}:00+05:00`).toISOString()
}

export function todayEkb(addDays = 0) {
  return ekbDate(new Date(Date.now() + addDays * DAY).toISOString())
}

export function dayLabel(iso: string) {
  const day = ekbDate(iso)
  if (day === todayEkb()) return 'сегодня'
  if (day === todayEkb(1)) return 'завтра'
  return new Date(iso)
    .toLocaleDateString('ru-RU', { weekday: 'short', day: 'numeric', month: 'short', timeZone: TZ })
    .replace('.', '')
}

export function greeting() {
  const hour = Number(ekbTime(new Date().toISOString()).slice(0, 2))
  if (hour >= 5 && hour < 12) return 'Доброе утро'
  if (hour >= 12 && hour < 18) return 'Добрый день'
  if (hour >= 18 && hour < 23) return 'Добрый вечер'
  return 'Доброй ночи'
}
