const TZ = 'Asia/Yekaterinburg'

export function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'long',
    timeZone: TZ,
  })
}

export function formatMeetingTime(iso: string) {
  const date = new Date(iso)
  const day = date.toLocaleDateString('ru-RU', {
    weekday: 'short',
    day: 'numeric',
    month: 'long',
    timeZone: TZ,
  })
  const time = date.toLocaleTimeString('ru-RU', {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: TZ,
  })
  return `${day[0].toUpperCase()}${day.slice(1)} · ${time}`
}

export function minutesBetween(start: string, end?: string) {
  return end ? Math.round((Date.parse(end) - Date.parse(start)) / 60000) : null
}

export function isPast(iso: string) {
  return Date.parse(iso) < Date.now()
}
