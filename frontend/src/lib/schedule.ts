import type { AvailabilitySlots } from '../api/profiles'

const DAYS = [
  ['mon', 'Пн'],
  ['tue', 'Вт'],
  ['wed', 'Ср'],
  ['thu', 'Чт'],
  ['fri', 'Пт'],
  ['sat', 'Сб'],
  ['sun', 'Вс'],
] as const

/** [19, 20, 22] → «19–21, 22–23» */
function toRanges(hours: number[]) {
  const sorted = [...hours].sort((a, b) => a - b)
  const ranges: string[] = []
  let start = sorted[0]
  sorted.forEach((hour, index) => {
    if (sorted[index + 1] !== hour + 1) {
      ranges.push(`${start}–${hour + 1}`)
      start = sorted[index + 1]
    }
  })
  return ranges.join(', ')
}

/** { tue: [19, 20], thu: [19, 20], sat: [10, 11] } → «Вт, Чт 19–21 · Сб 10–12» */
export function summarizeSlots(slots?: AvailabilitySlots) {
  if (!slots) return ''
  const groups = new Map<string, string[]>()
  DAYS.forEach(([key, label]) => {
    const hours = slots[key] ?? []
    if (hours.length === 0) return
    const ranges = toRanges(hours)
    groups.set(ranges, [...(groups.get(ranges) ?? []), label])
  })
  return [...groups].map(([ranges, days]) => `${days.join(', ')} ${ranges}`).join(' · ')
}
