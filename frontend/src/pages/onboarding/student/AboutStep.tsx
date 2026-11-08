import { Input } from '../../../components/ui/Input'
import { Segmented } from '../../../components/ui/Segmented'
import { Select } from '../../../components/ui/Select'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'

const STATUSES = [
  { value: 'student', label: 'Студент' },
  { value: 'graduate', label: 'Выпускник' },
  { value: 'junior', label: 'Молодой специалист' },
  { value: 'switcher', label: 'Меняю профессию' },
]

const UNIVERSITIES = [
  { value: 'urfu', label: 'УрФУ' },
  { value: 'usue', label: 'УрГЭУ' },
  { value: 'usmu', label: 'УГМУ' },
  { value: 'uspu', label: 'УрГПУ' },
  { value: 'usla', label: 'УрГЮУ' },
  { value: 'other', label: 'Другой' },
]

const COURSES = ['1', '2', '3', '4', '5', '6'].map((n) => ({ value: n, label: `${n} курс` }))

const EXPERIENCE = [
  { value: 'none', label: 'Нет опыта' },
  { value: 'under_1', label: 'До 1 года' },
  { value: '1_to_3', label: '1–3 года' },
  { value: 'over_3', label: '3+ года' },
]

type AboutStepProps = {
  value: StudentAnswers
  update: (patch: Partial<StudentAnswers>) => void
}

export function AboutStep({ value, update }: AboutStepProps) {
  const studies = value.status === 'student' || value.status === 'graduate'
  const works = value.status === 'junior' || value.status === 'switcher'

  return (
    <>
      <div className={styles.grid3}>
        <Input
          label="Фамилия"
          value={value.lastName}
          onChange={(e) => update({ lastName: e.target.value })}
        />
        <Input
          label="Имя"
          value={value.firstName}
          onChange={(e) => update({ firstName: e.target.value })}
        />
        <Input
          label="Отчество"
          placeholder="Если есть"
          value={value.middleName}
          onChange={(e) => update({ middleName: e.target.value })}
        />
      </div>

      <div className={styles.grid3}>
        <Input
          label="Возраст"
          type="number"
          min={14}
          max={99}
          value={value.age}
          onChange={(e) => update({ age: e.target.value })}
        />
      </div>

      <div className={styles.field}>
        <span className={styles.label}>Сейчас я</span>
        <Segmented
          options={STATUSES}
          value={value.status}
          onChange={(status) => update({ status })}
        />
      </div>

      {studies ? (
        <div className={styles.grid2}>
          <Select
            label="Вуз"
            placeholder="Выбери вуз"
            options={UNIVERSITIES}
            value={value.university}
            onChange={(e) => update({ university: e.target.value })}
          />
          <Input
            label="Направление подготовки"
            placeholder="Например, Программная инженерия"
            value={value.direction}
            onChange={(e) => update({ direction: e.target.value })}
          />
        </div>
      ) : null}

      {value.status === 'student' ? (
        <div className={styles.grid3}>
          <Select
            label="Курс"
            placeholder="Выбери курс"
            options={COURSES}
            value={value.course}
            onChange={(e) => update({ course: e.target.value })}
          />
        </div>
      ) : null}

      {works ? (
        <div className={styles.grid2}>
          <Input
            label={value.status === 'switcher' ? 'Кем работаешь сейчас' : 'Должность'}
            placeholder={value.status === 'switcher' ? 'Например, бухгалтер' : 'Например, Junior-разработчик'}
            value={value.position}
            onChange={(e) => update({ position: e.target.value })}
          />
          <Input
            label="Компания"
            placeholder="Необязательно"
            value={value.company}
            onChange={(e) => update({ company: e.target.value })}
          />
        </div>
      ) : null}

      <div className={styles.field}>
        <span className={styles.label}>Опыт работы в сфере</span>
        <Segmented
          options={EXPERIENCE}
          value={value.experience}
          onChange={(experience) => update({ experience })}
        />
      </div>
    </>
  )
}