import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { HourGrid } from '../../../components/ui/HourGrid'
import { Select } from '../../../components/ui/Select'
import { Slider } from '../../../components/ui/Slider'
import { plural } from '../../../lib/plural'
import { TIMEZONES } from '../../../lib/timezones'
import { COMMUNICATION_STYLES } from '../styles'
import type { MentorAnswers } from '../types'
import styles from '../Onboarding.module.css'

type TimeStepProps = {
  value: MentorAnswers
  update: (patch: Partial<MentorAnswers>) => void
}

export function TimeStep({ value, update }: TimeStepProps) {
  return (
    <>
      <Slider
        label="Время на наставничество в неделю"
        min={1}
        max={10}
        value={value.hoursPerWeek}
        onChange={(hoursPerWeek) => update({ hoursPerWeek })}
        format={(n) => (n === 10 ? '10+ часов' : plural(n, ['час', 'часа', 'часов']))}
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
        label="Удобное время для встреч"
        value={value.schedule}
        onChange={(schedule) => update({ schedule })}
      />

      <div className={styles.field}>
        <span className={styles.label}>Мой стиль наставничества</span>
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

      <label className={styles.check}>
        <input
          type="checkbox"
          checked={value.visible}
          onChange={(e) => update({ visible: e.target.checked })}
        />
        <span className={styles.checkText}>
          <span className={styles.checkTitle}>Показывать анкету в каталоге после проверки</span>
          <span className={styles.hint}>Можно скрыть в любой момент, если нет свободных мест</span>
        </span>
      </label>
    </>
  )
}