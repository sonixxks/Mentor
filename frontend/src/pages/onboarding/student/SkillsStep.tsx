import { X } from 'lucide-react'
import { Segmented } from '../../../components/ui/Segmented'
import { Select } from '../../../components/ui/Select'
import { Tag } from '../../../components/ui/Tag'
import { SKILLS } from '../skills'
import type { SkillLevel, StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'

const LEVELS = [
  { value: 'beginner', label: 'Новичок' },
  { value: 'confident', label: 'Уверенно' },
  { value: 'pro', label: 'Профи' },
]

type SkillsStepProps = {
  value: StudentAnswers
  update: (patch: Partial<StudentAnswers>) => void
}

export function SkillsStep({ value, update }: SkillsStepProps) {
  const { targetSkills, currentSkills } = value

  function toggleTarget(name: string) {
    update({
      targetSkills: targetSkills.includes(name)
        ? targetSkills.filter((skill) => skill !== name)
        : [...targetSkills, name],
    })
  }

  function addCurrent(name: string) {
    if (!name) return
    update({ currentSkills: [...currentSkills, { name, level: 'beginner' }] })
  }

  function setLevel(name: string, level: SkillLevel) {
    update({
      currentSkills: currentSkills.map((skill) =>
        skill.name === name ? { ...skill, level } : skill,
      ),
    })
  }

  function removeCurrent(name: string) {
    update({ currentSkills: currentSkills.filter((skill) => skill.name !== name) })
  }

  const available = SKILLS.filter((name) => !currentSkills.some((skill) => skill.name === name))

  return (
    <>
      <div className={styles.field}>
        <span className={styles.label}>Хочу развить</span>
        <div className={styles.tags}>
          {SKILLS.map((name) => (
            <Tag
              key={name}
              selected={targetSkills.includes(name)}
              onClick={() => toggleTarget(name)}
            >
              {name}
            </Tag>
          ))}
        </div>
      </div>

      <div className={styles.field}>
        <span className={styles.label}>Уже умею</span>
        <div className={styles.list}>
          {currentSkills.map((skill) => (
            <div key={skill.name} className={styles.skillRow}>
              <span className={styles.skillName}>{skill.name}</span>
              <div className={styles.level}>
                <Segmented
                  options={LEVELS}
                  value={skill.level}
                  onChange={(level) => setLevel(skill.name, level as SkillLevel)}
                />
              </div>
              <button
                type="button"
                className={styles.remove}
                aria-label={`Удалить ${skill.name}`}
                onClick={() => removeCurrent(skill.name)}
              >
                <X size={18} />
              </button>
            </div>
          ))}

          {available.length > 0 ? (
            <div className={styles.addSkill}>
              <Select
                label="Добавить навык"
                placeholder="Выбери навык"
                options={available.map((name) => ({ value: name, label: name }))}
                value=""
                onChange={(e) => addCurrent(e.target.value)}
              />
            </div>
          ) : null}
        </div>
      </div>
    </>
  )
}
