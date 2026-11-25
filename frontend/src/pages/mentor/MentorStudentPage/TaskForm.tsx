import { useState } from 'react'
import type { FormEvent } from 'react'
import type { TaskInput } from '../../../api/workspaces'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Textarea } from '../../../components/ui/Textarea'
import { ekbDate, fromEkb, todayEkb } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import styles from './MentorStudentPage.module.css'

type TaskFormProps = {
  initial?: { title: string; description?: string; deadline_at: string }
  submitLabel: string
  pending: boolean
  error?: unknown
  onSubmit: (input: TaskInput) => void
  onCancel?: () => void
}

export function TaskForm({
  initial,
  submitLabel,
  pending,
  error,
  onSubmit,
  onCancel,
}: TaskFormProps) {
  const [title, setTitle] = useState(initial?.title ?? '')
  const [description, setDescription] = useState(initial?.description ?? '')
  const [deadline, setDeadline] = useState(initial ? ekbDate(initial.deadline_at) : todayEkb(7))

  const ready = Boolean(title.trim() && deadline)

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (!ready) return
    onSubmit({
      title: title.trim(),
      description: description.trim(),
      deadline_at: fromEkb(deadline, '23:59'),
    })
  }

  return (
    <form className={styles.form} onSubmit={handleSubmit}>
      <Input
        label="Название"
        placeholder="Например, обучить модель на датасете"
        maxLength={200}
        value={title}
        onChange={(e) => setTitle(e.target.value)}
      />
      <Textarea
        label="Описание"
        rows={3}
        maxLength={2000}
        placeholder="Что сделать и как понять, что готово"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />
      <Input
        label="Дедлайн"
        type="date"
        min={initial ? undefined : todayEkb()}
        value={deadline}
        onChange={(e) => setDeadline(e.target.value)}
      />

      {error ? (
        <p className={styles.error} role="alert">
          {errorMessage(error)}
        </p>
      ) : null}

      <div className={styles.formActions}>
        {onCancel ? (
          <Button variant="secondary" onClick={onCancel}>
            Отмена
          </Button>
        ) : null}
        <Button type="submit" variant="accent" fullWidth={!onCancel} disabled={!ready || pending}>
          {pending ? 'Сохраняем…' : submitLabel}
        </Button>
      </div>
    </form>
  )
}
