import { Input } from '../../../components/ui/Input'
import { Segmented } from '../../../components/ui/Segmented'
import { Slider } from '../../../components/ui/Slider'
import { Tag } from '../../../components/ui/Tag'
import { plural } from '../../../lib/plural'
import type { MentorAnswers } from '../types'
import styles from '../Onboarding.module.css'
import { MAX_MONTHS } from '../types'

const LEVELS = [
  { value: 'beginner', label: 'Начинающие' },
  { value: 'base', label: 'Есть база' },
  { value: 'advanced', label: 'Продвинутые' },
]

const PRICE = [
  { value: 'free', label: 'Бесплатно' },
  { value: 'paid', label: 'Платно' },
]

function toggle(list: string[], item: string) {
  return list.includes(item) ? list.filter((x) => x !== item) : [...list, item]
}

type FormatStepProps = {
  value: MentorAnswers
  update: (patch: Partial<MentorAnswers>) => void
}

export function FormatStep({ value, update }: FormatStepProps) {
  return (
    <>
      <div className={styles.field}>
        <span className={styles.label}>Уровень учеников · можно несколько</span>
        <div className={styles.tags}>
          {LEVELS.map((level) => (
            <Tag
              key={level.value}
              selected={value.studentLevels.includes(level.value)}
              onClick={() => update({ studentLevels: toggle(value.studentLevels, level.value) })}
            >
              {level.label}
            </Tag>
          ))}
        </div>
      </div>

      <div className={styles.field}>
        <span className={styles.label}>Длительность менторства, месяцев</span>
        <div className={styles.grid3}>
          <Input
            label="От"
            type="number"
            min={1}
            max={MAX_MONTHS}
            value={value.durationFrom || ''}
            onChange={(e) => update({ durationFrom: Number(e.target.value) })}
          />
          <Input
            label="До"
            type="number"
            min={1}
            max={MAX_MONTHS}
            value={value.durationTo || ''}
            onChange={(e) => update({ durationTo: Number(e.target.value) })}
          />
        </div>
      </div>

      <Slider
        label="Сколько учеников одновременно"
        min={1}
        max={10}
        value={value.maxStudents}
        onChange={(maxStudents) => update({ maxStudents })}
        format={(n) => plural(n, ['ученик', 'ученика', 'учеников'])}
      />

      <div className={styles.field}>
        <span className={styles.label}>Стоимость занятия</span>
        <Segmented
          options={PRICE}
          value={value.paid ? 'paid' : 'free'}
          onChange={(price) => update({ paid: price === 'paid' })}
        />
      </div>

      {value.paid ? (
        <div className={styles.field}>
          <div className={styles.grid3}>
            <Input
              label="Цена за занятие 60 минут, ₽"
              type="number"
              min={0}
              placeholder="Например, 1500"
              value={value.price}
              onChange={(e) => update({ price: e.target.value })}
            />
          </div>
          <p className={styles.hint}>
            Платформа не принимает оплату: о расчётах вы договариваетесь с учеником напрямую.
          </p>
        </div>
      ) : null}
    </>
  )
}