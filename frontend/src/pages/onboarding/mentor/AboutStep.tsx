import { FileUpload } from '../../../components/ui/FileUpload'
import { Input } from '../../../components/ui/Input'
import { Segmented } from '../../../components/ui/Segmented'
import { Textarea } from '../../../components/ui/Textarea'
import type { MentorAnswers } from '../types'
import styles from '../Onboarding.module.css'
import { X } from 'lucide-react'

const EXPERIENCE = [
  { value: '1_to_3', label: '1–3 года' },
  { value: '3_to_5', label: '3–5 лет' },
  { value: '5_to_10', label: '5–10 лет' },
  { value: 'over_10', label: '10+ лет' },
]

type AboutStepProps = {
  value: MentorAnswers
  update: (patch: Partial<MentorAnswers>) => void
  documents: File[]
  onDocumentsChange: (files: File[]) => void
}

export function AboutStep({ value, update, documents, onDocumentsChange }: AboutStepProps) {
  const { links } = value

  function setLink(index: number, link: string) {
    update({ links: links.map((item, i) => (i === index ? link : item)) })
  }

  function addLink() {
    update({ links: [...links, ''] })
  }

  function removeLink(index: number) {
    const rest = links.filter((_, i) => i !== index)
    update({ links: rest.length > 0 ? rest : [''] })
  }

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

      <div className={styles.grid2}>
        <Input
          label="Должность"
          placeholder="Например, ML-инженер"
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

      <div className={styles.field}>
        <span className={styles.label}>Опыт работы в сфере</span>
        <Segmented
          options={EXPERIENCE}
          value={value.experience}
          onChange={(experience) => update({ experience })}
        />
      </div>

      <Textarea
        label="Коротко о себе"
        placeholder="Например, помогаю начинающим ML-инженерам собрать первый проект и пройти собеседования"
        maxLength={500}
        value={value.about}
        onChange={(e) => update({ about: e.target.value })}
      />

      <FileUpload
        label="Подтверждение опыта"
        hint="Диплом, сертификаты или справка с работы. PDF, JPG или PNG до 10 МБ. Видит только куратор."
        accept=".pdf,.jpg,.jpeg,.png"
        files={documents}
        onChange={onDocumentsChange}
      />

      <div className={styles.field}>
        <span className={styles.label}>Ссылки · GitHub, LinkedIn, портфолио</span>
        <div className={styles.list}>
          {links.map((link, index) => (
            <div key={index} className={styles.linkRow}>
              <Input
                label={`Ссылка ${index + 1}`}
                placeholder="github.com/username"
                value={link}
                onChange={(e) => setLink(index, e.target.value)}
              />
              <button
                type="button"
                className={styles.remove}
                aria-label={`Удалить ссылку ${index + 1}`}
                onClick={() => removeLink(index)}
              >
                <X size={18} />
              </button>
            </div>
          ))}
          <button type="button" className={styles.addLink} onClick={addLink}>
            + Добавить ссылку
          </button>
        </div>
      </div>
    </>
  )
}
