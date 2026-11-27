import { X } from 'lucide-react'
import { Segmented } from '../../../components/ui/Segmented'
import { Select } from '../../../components/ui/Select'
import { Tag } from '../../../components/ui/Tag'
import { useSkills } from '../../../hooks/useSkills'
import type { SkillLevel, StudentAnswers } from '../types'
import styles from '../Onboarding.module.css'

const LEVELS = [
  { value: '1', label: 'Новичок' },
  { value: '2', label: 'Уверенно' },
  { value: '3', label: 'Профи' },
]

type SkillsStepProps = {
  value: StudentAnswers
  update: (patch: Partial<StudentAnswers>) => void
}

export function SkillsStep({ value, update }: SkillsStepProps) {
  const skillsQuery = useSkills()
  const allSkills = skillsQuery.data ?? []

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
    update({ currentSkills: [...currentSkills, { name, level: 1 }] })
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

  const available = allSkills.filter(
    (name) => !currentSkills.some((skill) => skill.name === name),
  )

  if (skillsQuery.isPending) {
    return <p className={styles.hint}>Загружаем список навыков...</p>
  }

  if (skillsQuery.isError) {
    return <p className={styles.error}>Не удалось загрузить список навыков</p>
  }

  return (
    <>
      <div className={styles.field}>
        <span className={styles.label}>Хочу развить</span>
        <div className={styles.tags}>
          {allSkills.map((name) => (
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
                  value={String(skill.level)}
                  onChange={(level) => setLevel(skill.name, Number(level) as SkillLevel)}
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