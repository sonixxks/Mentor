
import { useEffect, useState } from 'react'
import { X } from 'lucide-react'
import { Segmented } from '../../../components/ui/Segmented'
import { Select } from '../../../components/ui/Select'
import { Tag } from '../../../components/ui/Tag'
import { getSkills } from '../../../api/profiles'
import type { MentorAnswers, MentorSkill } from '../types'
import styles from '../Onboarding.module.css'

const LEVELS = [
  { value: 'confident', label: 'Уверенно' },
  { value: 'expert', label: 'Эксперт' },
]

const HELP_TOPICS = [
  { value: 'interview_prep', title: 'Подготовка к собеседованию' },
  { value: 'code_review', title: 'Разбор кода' },
  { value: 'portfolio', title: 'Портфолио и проекты' },
  { value: 'career_growth', title: 'Развитие карьеры' },
  { value: 'change_direction', title: 'Смена направления' },
  { value: 'soft_skills', title: 'Развитие soft skills' },
  { value: 'first_job_in_it', title: 'Первая работа в IT' },
]

type SkillsStepProps = {
  value: MentorAnswers
  update: (patch: Partial<MentorAnswers>) => void
}

export function SkillsStep({ value, update }: SkillsStepProps) {
  const [availableSkills, setAvailableSkills] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState('')

  const { skills, goals } = value

  useEffect(() => {
    let cancelled = false

    async function loadSkills() {
      try {
        const result = await getSkills()

        if (!cancelled) {
          setAvailableSkills(result)
          setLoadError('')
        }
      } catch {
        if (!cancelled) {
          setLoadError('Не удалось загрузить список навыков с сервера')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    loadSkills()

    return () => {
      cancelled = true
    }
  }, [])

  function addSkill(name: string) {
    if (!name || skills.some((skill) => skill.name === name)) return

    update({
      skills: [...skills, { name, level: 'confident' }],
    })
  }

  function setLevel(name: string, level: MentorSkill['level']) {
    update({
      skills: skills.map((skill) =>
        skill.name === name ? { ...skill, level } : skill,
      ),
    })
  }

  function removeSkill(name: string) {
    update({
      skills: skills.filter((skill) => skill.name !== name),
    })
  }

  function toggleGoal(goal: string) {
    update({
      goals: goals.includes(goal)
        ? goals.filter((item) => item !== goal)
        : [...goals, goal],
    })
  }

  const available = availableSkills.filter(
    (name) => !skills.some((skill) => skill.name === name),
  )

  return (
    <>
      <div className={styles.field}>
        <span className={styles.label}>Навыки, которыми делюсь</span>

        <div className={styles.list}>
          {skills.map((skill) => (
            <div key={skill.name} className={styles.skillRow}>
              <span className={styles.skillName}>{skill.name}</span>

              <div className={styles.level}>
                <Segmented
                  options={LEVELS}
                  value={skill.level}
                  onChange={(level) =>
                    setLevel(skill.name, level as MentorSkill['level'])
                  }
                />
              </div>

              <button
                type="button"
                className={styles.remove}
                aria-label={`Удалить ${skill.name}`}
                onClick={() => removeSkill(skill.name)}
              >
                <X size={18} />
              </button>
            </div>
          ))}

          {loading ? (
            <p className={styles.hint}>Загружаем список навыков...</p>
          ) : null}

          {loadError ? (
            <p className={styles.error}>{loadError}</p>
          ) : null}

          {!loading && !loadError && available.length > 0 ? (
            <div className={styles.addSkill}>
              <Select
                label="Добавить навык"
                placeholder="Выберите навык"
                options={available.map((name) => ({
                  value: name,
                  label: name,
                }))}
                value=""
                onChange={(e) => addSkill(e.target.value)}
              />
            </div>
          ) : null}

          {!loading && !loadError && available.length === 0 && skills.length > 0 ? (
            <p className={styles.hint}>Все доступные навыки уже добавлены</p>
          ) : null}
        </div>
      </div>

      <div className={styles.field}>
        <span className={styles.label}>С чем помогаю</span>

        <div className={styles.tags}>
          {HELP_TOPICS.map((topic) => (
            <Tag
              key={topic.value}
              selected={goals.includes(topic.value)}
              onClick={() => toggleGoal(topic.value)}
            >
              {topic.title}
            </Tag>
          ))}
        </div>
      </div>
    </>
  )
}
