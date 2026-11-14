import { clsx } from 'clsx'
import styles from './Steps.module.css'

type StepsProps = {
  steps: string[]
  current: number
}

export function Steps({ steps, current }: StepsProps) {
  return (
    <ol className={styles.steps} style={{ gridTemplateColumns: `repeat(${steps.length}, minmax(0, 1fr))` }}>
      {steps.map((name, index) => {
        const number = index + 1
        return (
          <li
            key={name}
            className={clsx(
              styles.step,
              number < current ? styles.done : '',
              number === current ? styles.current : '',
            )}
            aria-current={number === current ? 'step' : undefined}
          >
            <span className={styles.bar} />
            <span>
              {number} <span className={styles.name}>· {name}</span>
            </span>
          </li>
        )
      })}
    </ol>
  )
}