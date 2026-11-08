import { X } from 'lucide-react'
import { Segmented } from '../../../components/ui/Segmented'
import { Select } from '../../../components/ui/Select'
import { Tag } from '../../../components/ui/Tag'
import { SKILLS, TOPICS } from '../skills'
import type { MentorAnswers, MentorSkill } from '../types'
import styles from '../Onboarding.module.css'

const LEVELS = [
  { value: 'confident', label: 'Уверенно' },
  { value: 'expert', label: 'Эксперт' },
]

type SkillsStepProps = {
  value: MentorAnswers
  update: (patch: Partial<MentorAnswers>) => void
}

export function SkillsStep({ value, update }: SkillsStepProps) {
  const { skills, topics } = value

  function addSkill(name: string) {
    if (!name) return
    update({ skills: [...skills, { name, level: 'confident' }] })
  }

  function setLevel(name: string, level: MentorSkill['level']) {
    update({
      skills: skills.map((skill) => (skill.name === name ? { ...skill, level } : skill)),
    })
  }

  function removeSkill(name: string) {
    update({ skills: skills.filter((skill) => skill.name !== name) })
  }

  function toggleTopic(topic: string) {
    update({
      topics: topics.includes(topic)
        ? topics.filter((item) => item !== topic)
        : [...topics, topic],
    })
  }

  const available = SKILLS.filter((name) => !skills.some((skill) => skill.name === name))

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
                  onChange={(level) => setLevel(skill.name, level as MentorSkill['level'])}
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

          {available.length > 0 ? (
            <div className={styles.addSkill}>
              <Select
                label="Добавить навык"
                placeholder="Выберите навык"
                options={available.map((name) => ({ value: name, label: name }))}
                value=""
                onChange={(e) => addSkill(e.target.value)}
              />
            </div>
          ) : null}
        </div>
      </div>

      <div className={styles.field}>
        <span className={styles.label}>С чем ещё помогаю</span>
        <div className={styles.tags}>
          {TOPICS.map((topic) => (
            <Tag key={topic} selected={topics.includes(topic)} onClick={() => toggleTopic(topic)}>
              {topic}
            </Tag>
          ))}
        </div>
      </div>
    </>
  )
}