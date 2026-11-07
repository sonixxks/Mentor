import type { CSSProperties } from 'react'
import styles from './Slider.module.css'

type SliderProps = {
  label: string
  value: number
  min: number
  max: number
  step?: number
  onChange: (value: number) => void
  format?: (value: number) => string
  marks?: string[]
}

export function Slider({ label, value, min, max, step = 1, onChange, format, marks }: SliderProps) {
  const fill = ((value - min) / (max - min)) * 100

  return (
    <div className={styles.field}>
      <div className={styles.head}>
        <span className={styles.label}>{label}</span>
        <span className={styles.value}>{format ? format(value) : value}</span>
      </div>
      <input
        type="range"
        aria-label={label}
        className={styles.range}
        style={{ '--fill': `${fill}%` } as CSSProperties}
        min={min}
        max={max}
        step={step}
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
      />
      {marks ? (
        <div className={styles.marks}>
          {marks.map((mark) => (
            <span key={mark}>{mark}</span>
          ))}
        </div>
      ) : null}
    </div>
  )
}