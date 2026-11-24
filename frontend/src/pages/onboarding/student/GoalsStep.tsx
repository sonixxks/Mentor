import { ChoiceCard } from '../../../components/ui/ChoiceCard'
import { Input } from '../../../components/ui/Input'
import { Textarea } from '../../../components/ui/Textarea'
import { GOALS } from '../goals'
import { MAX_MONTHS } from '../types'
import type { StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'

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

      <div className={styles.field}>
        <span className={styles.label}>Срок программы, месяцев</span>
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
