import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { Slider } from '../../../components/ui/Slider'
import { Textarea } from '../../../components/ui/Textarea'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'
import { plural } from '../../../lib/plural'
import { GOALS } from '../goals'
import { MAX_MONTHS } from '../types'

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
        max={MAX_MONTHS}
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
