import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { HourGrid } from '../../../components/ui/HourGrid'
import { Select } from '../../../components/ui/Select'
import { Slider } from '../../../components/ui/Slider'
import { plural } from '../../../lib/plural'
import { TIMEZONES } from '../../../lib/timezones'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'

const STYLES = [
  { value: 'friendly', title: 'Дружелюбный', description: 'Неформально, поддержка и мотивация' },
  { value: 'structured', title: 'Структурный', description: 'Чёткий план, дедлайны, отчёты' },
  { value: 'free', title: 'Свободный', description: 'Встречи по запросу, без жёстких рамок' },
]

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

      <div className={styles.grid2}>
        <Select
          label="Часовой пояс"
          options={TIMEZONES}
          value={value.timezone}
          onChange={(e) => update({ timezone: e.target.value })}
        />
      </div>

      <HourGrid
        label="Удобное время"
        value={value.schedule}
        onChange={(schedule) => update({ schedule })}
      />

      <div className={styles.field}>
        <span className={styles.label}>Стиль общения</span>
        <div className={styles.grid3}>
          {STYLES.map((item) => (
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