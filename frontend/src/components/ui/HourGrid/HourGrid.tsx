import { useEffect, useRef } from 'react'
import { clsx } from 'clsx'
import styles from './HourGrid.module.css'

export const DAYS = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'] as const
export type Day = (typeof DAYS)[number]
export type Schedule = Record<Day, number[]>

const DAY_LABELS: Record<Day, string> = {
  mon: 'Пн',
  tue: 'Вт',
  wed: 'Ср',
  thu: 'Чт',
  fri: 'Пт',
  sat: 'Сб',
  sun: 'Вс',
}

export const emptySchedule = (): Schedule => ({
  mon: [],
  tue: [],
  wed: [],
  thu: [],
  fri: [],
  sat: [],
  sun: [],
})

function toRanges(hours: number[]) {
  const sorted = [...hours].sort((a, b) => a - b)
  const ranges: string[] = []
  let start = sorted[0]
  for (let i = 0; i < sorted.length; i++) {
    if (sorted[i + 1] !== sorted[i] + 1) {
      ranges.push(`${start}–${sorted[i] + 1}`)
      start = sorted[i + 1]
    }
  }
  return ranges.join(', ')
}

function summarize(schedule: Schedule) {
  return DAYS.filter((day) => schedule[day].length > 0)
    .map((day) => `${DAY_LABELS[day]} ${toRanges(schedule[day])}`)
    .join(' · ')
}

type HourGridProps = {
  label: string
  value: Schedule
  onChange: (value: Schedule) => void
  from?: number
  to?: number
}

export function HourGrid({ label, value, onChange, from = 8, to = 23 }: HourGridProps) {
  const hours = Array.from({ length: to - from }, (_, i) => from + i)
  const painting = useRef<'add' | 'remove' | null>(null)

  useEffect(() => {
    const stop = () => {
      painting.current = null
    }
    window.addEventListener('pointerup', stop)
    return () => window.removeEventListener('pointerup', stop)
  }, [])

  function apply(day: Day, hour: number, mode: 'add' | 'remove') {
    const has = value[day].includes(hour)
    if (mode === 'add' && !has) onChange({ ...value, [day]: [...value[day], hour] })
    if (mode === 'remove' && has)
      onChange({ ...value, [day]: value[day].filter((h) => h !== hour) })
  }

  function handleDown(day: Day, hour: number) {
    const mode = value[day].includes(hour) ? 'remove' : 'add'
    painting.current = mode
    apply(day, hour, mode)
  }

  function handleEnter(day: Day, hour: number) {
    if (painting.current) apply(day, hour, painting.current)
  }

  const summary = summarize(value)

  return (
    <div className={styles.field}>
      <div className={styles.head}>
        <span className={styles.label}>
          {label}
        </span>
        <span className={styles.summary}>{summary || 'Нажмите или протяните'}</span>
      </div>

      <div
        className={styles.grid}
        style={{ gridTemplateColumns: `44px repeat(${DAYS.length}, minmax(0, 1fr))` }}
      >
        <span />
        {DAYS.map((day) => (
          <span key={day} className={styles.dayHead}>
            {DAY_LABELS[day]}
          </span>
        ))}

        {hours.map((hour) => (
          <div key={hour} style={{ display: 'contents' }}>
            <span className={styles.hour}>{hour}:00</span>
            {DAYS.map((day) => {
              const on = value[day].includes(hour)
              return (
                <button
                  key={day}
                  type="button"
                  aria-label={`${DAY_LABELS[day]}, ${hour}:00–${hour + 1}:00`}
                  aria-pressed={on}
                  className={clsx(styles.cell, on ? styles.on : '')}
                  onPointerDown={() => handleDown(day, hour)}
                  onPointerEnter={() => handleEnter(day, hour)}
                />
              )
            })}
          </div>
        ))}
      </div>
    </div>
  )
}
