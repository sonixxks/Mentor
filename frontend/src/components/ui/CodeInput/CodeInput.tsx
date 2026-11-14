import { useRef } from 'react'
import type { ChangeEvent, ClipboardEvent, KeyboardEvent } from 'react'
import { clsx } from 'clsx'
import styles from './CodeInput.module.css'

type CodeInputProps = {
  value: string
  onChange: (value: string) => void
  length?: number
  error?: string
}

export function CodeInput({ value, onChange, length = 6, error }: CodeInputProps) {
  const inputs = useRef<Array<HTMLInputElement | null>>([])
  const digits = Array.from({ length }, (_, i) => value[i] ?? '')

  function handleChange(index: number, event: ChangeEvent<HTMLInputElement>) {
    const digit = event.target.value.replace(/\D/g, '').slice(-1)
    if (!digit) return
    onChange((value.slice(0, index) + digit).slice(0, length))
    inputs.current[index + 1]?.focus()
  }

  function handleKeyDown(index: number, event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== 'Backspace') return
    event.preventDefault()
    if (digits[index]) {
      onChange(value.slice(0, index))
    } else if (index > 0) {
      onChange(value.slice(0, index - 1))
      inputs.current[index - 1]?.focus()
    }
  }

  function handlePaste(event: ClipboardEvent<HTMLInputElement>) {
    const pasted = event.clipboardData.getData('text').replace(/\D/g, '').slice(0, length)
    if (!pasted) return
    event.preventDefault()
    onChange(pasted)
    inputs.current[Math.min(pasted.length, length - 1)]?.focus()
  }

  return (
    <div className={styles.wrap}>
      <div className={styles.cells}>
        {digits.map((digit, index) => (
          <input
            key={index}
            ref={(element) => {
              inputs.current[index] = element
            }}
            className={clsx(styles.cell, error ? styles.invalid : '')}
            value={digit}
            inputMode="numeric"
            autoComplete={index === 0 ? 'one-time-code' : 'off'}
            aria-label={`Цифра ${index + 1}`}
            onChange={(event) => handleChange(index, event)}
            onKeyDown={(event) => handleKeyDown(index, event)}
            onPaste={handlePaste}
            onFocus={(event) => event.target.select()}
          />
        ))}
      </div>
      {error ? <span className={styles.error}>{error}</span> : null}
    </div>
  )
}
