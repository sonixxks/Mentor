import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { Slider } from '../../../components/ui/Slider'
import { Textarea } from '../../../components/ui/Textarea'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'
import { plural } from '../../../lib/plural'

const GOALS = [
  {
    value: 'job',
    title: 'Найти стажировку или работу',
    description: 'Резюме, портфолио, собеседования',
  },
  { value: 'skill', title: 'Освоить новый навык', description: 'С нуля до уверенного уровня' },
  { value: 'project', title: 'Сделать проект', description: 'Пет-проект или диплом с поддержкой' },
  {
    value: 'interview',
    title: 'Подготовиться к собеседованию',
    description: 'Технические и HR-раунды',
  },
  { value: 'switch', title: 'Сменить направление', description: 'Понять, куда расти дальше' },
  { value: 'soft', title: 'Развить soft skills', description: 'Коммуникация, тайм-менеджмент' },
]

const MAX_GOALS = 2

type GoalsStepProps = {
  value: StudentAnswers
  update: (patch: Partial<StudentAnswers>) => void
}

export function GoalsStep({ value, update }: GoalsStepProps) {
  const { goals } = value

  function toggleGoal(goal: string) {
    if (goals.includes(goal)) {
      update({ goals: goals.filter((item) => item !== goal) })
    } else if (goals.length < MAX_GOALS) {
      update({ goals: [...goals, goal] })
    } else {
      update({ goals: [goals[1], goal] })
    }
  }

  return (
    <>
      <div className={styles.field}>
        <span className={styles.label}>Главная цель · можно выбрать до двух</span>
        <div className={styles.grid2}>
          {GOALS.map((goal) => (
            <ChoiceCard
              key={goal.value}
              title={goal.title}
              description={goal.description}
              selected={goals.includes(goal.value)}
              onClick={() => toggleGoal(goal.value)}
            />
          ))}
        </div>
      </div>

      <Slider
        label="Срок программы"
        min={1}
        max={12}
        value={value.durationMonths}
        onChange={(durationMonths) => update({ durationMonths })}
        format={(n) => plural(n, ['месяц', 'месяца', 'месяцев'])}
      />

      <Textarea
        label="Какой результат будет успехом"
        placeholder="Например, проект по ML в портфолио и приглашение на стажировку к лету"
        maxLength={300}
        value={value.success}
        onChange={(e) => update({ success: e.target.value })}
      />
    </>
  )
}
