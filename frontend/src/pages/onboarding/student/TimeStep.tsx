import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { HourGrid } from '../../../components/ui/HourGrid'
import { Slider } from '../../../components/ui/Slider'
import { plural } from '../../../lib/plural'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'
import { COMMUNICATION_STYLES } from '../styles'

type TimeStepProps = {
  value: StudentAnswers
  update: (patch: Partial<StudentAnswers>) => void
}

export function TimeStep({ value, update }: TimeStepProps) {
  return (
    <>
      <Slider
        label="Сколько времени в неделю"
        min={1}
        max={8}
        value={value.hoursPerWeek}
        onChange={(hoursPerWeek) => update({ hoursPerWeek })}
        format={(n) => (n === 8 ? '8+ часов' : plural(n, ['час', 'часа', 'часов']))}
      />

      <HourGrid
        label="Удобное время"
        value={value.schedule}
        onChange={(schedule) => update({ schedule })}
      />

      <p className={styles.hint}>Время по Екатеринбургу (UTC+5)</p>

      <div className={styles.field}>
        <span className={styles.label}>Стиль общения</span>
        <div className={styles.grid3}>
          {COMMUNICATION_STYLES.map((item) => (
            <ChoiceCard
              key={item.value}
              title={item.title}
              description={item.description}
              selected={value.style === item.value}
              onClick={() => update({ style: item.value })}
            />
          ))}
        </div>
      </div>
    </>
  )
}
