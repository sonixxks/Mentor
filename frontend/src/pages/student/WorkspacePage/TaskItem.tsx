import { useMutation, useQueryClient } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { Check, ChevronDown, Link as LinkIcon } from 'lucide-react'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { changeTaskStatus } from '../../../api/workspaces'
import type { Task, TaskStatusUpdate } from '../../../api/workspaces'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { formatDate, isPast } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import { TASK_STATUS } from '../../../lib/labels'
import { plural } from '../../../lib/plural'
import { TaskComments } from './TaskComments'
import styles from './WorkspacePage.module.css'

type TaskItemProps = {
  task: Task
  workspaceId: string
}

function taskMeta(task: Task) {
  const parts: string[] = []
  if (task.status === 'completed' && task.completed_at) {
    parts.push(`выполнено ${formatDate(task.completed_at)}`)
  } else if (task.status === 'on_review') {
    parts.push('ждёт проверки наставника')
  } else {
    parts.push(`до ${formatDate(task.deadline_at)}`)
  }
  if (task.comments_count) {
    parts.push(plural(task.comments_count, ['комментарий', 'комментария', 'комментариев']))
  }
  return parts.join(' · ')
}

function isUrl(value: string) {
  try {
    new URL(value)
    return true
  } catch {
    return false
  }
}

export function TaskItem({ task, workspaceId }: TaskItemProps) {
  const queryClient = useQueryClient()
  const canSubmit = task.status === 'in_progress' || task.status === 'needs_revision'
  const [open, setOpen] = useState(canSubmit)
  const [link, setLink] = useState(task.artifacts_url ?? '')
  const [linkError, setLinkError] = useState('')

  const status = TASK_STATUS[task.status]
  const overdue =
    isPast(task.deadline_at) && task.status !== 'completed' && task.status !== 'on_review'

  const change = useMutation({
    mutationFn: (update: TaskStatusUpdate) => changeTaskStatus(task.id, update),
    onSuccess: (updated) => {
      queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId] })
      queryClient.invalidateQueries({ queryKey: ['workspaces'] })
      if (updated.status === 'in_progress') setOpen(true)
    },
  })

  function toggle() {
    setOpen((value) => !value)
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const url = link.trim()
    if (url && !isUrl(url)) {
      setLinkError('Ссылка должна начинаться с https://')
      return
    }
    setLinkError('')
    change.mutate({ status: 'on_review', artifacts_url: url || undefined })
  }

  return (
    <article className={clsx(styles.task, canSubmit ? styles.taskActive : '')}>
      <div className={styles.taskHead}>
        <span
          className={clsx(styles.check, task.status === 'completed' ? styles.checkDone : '')}
          aria-hidden="true"
        >
          {task.status === 'completed' ? <Check size={14} strokeWidth={3} /> : null}
        </span>

        <button type="button" className={styles.taskTitle} aria-expanded={open} onClick={toggle}>
          <span className={styles.taskName}>{task.title}</span>
          <span className={clsx(styles.taskMeta, overdue ? styles.overdue : '')}>
            {overdue ? 'просрочено · ' : ''}
            {taskMeta(task)}
          </span>
        </button>

        <Badge variant={status.variant}>{status.label}</Badge>

        {task.status === 'todo' ? (
          <Button
            variant="secondary"
            size="sm"
            disabled={change.isPending}
            onClick={() => change.mutate({ status: 'in_progress' })}
          >
            Начать
          </Button>
        ) : null}

        <button
          type="button"
          className={styles.toggle}
          aria-label={open ? 'Свернуть задачу' : 'Развернуть задачу'}
          aria-expanded={open}
          onClick={toggle}
        >
          <ChevronDown size={18} className={clsx(styles.chevron, open ? styles.chevronOpen : '')} />
        </button>
      </div>

      {open ? (
        <div className={styles.taskBody}>
          {task.description ? <p className={styles.text}>{task.description}</p> : null}

          {task.mentor_comment ? (
            <p className={styles.mentorComment}>
              <b>{task.status === 'needs_revision' ? 'Что доработать:' : 'Отзыв наставника:'}</b>{' '}
              {task.mentor_comment}
            </p>
          ) : null}

          {canSubmit ? (
            <form className={styles.submitForm} onSubmit={handleSubmit} noValidate>
              <label className={styles.fieldLabel} htmlFor={`link-${task.id}`}>
                Результат работы · ссылка на репозиторий, макет или документ, если есть
              </label>
              <div className={styles.submitRow}>
                <Input
                  id={`link-${task.id}`}
                  type="url"
                  placeholder="https://github.com/…"
                  value={link}
                  error={linkError}
                  onChange={(e) => setLink(e.target.value)}
                />
                <Button type="submit" size="md" disabled={change.isPending}>
                  {change.isPending ? 'Отправляем…' : 'Отправить на проверку'}
                </Button>
              </div>
            </form>
          ) : null}

          {!canSubmit && task.artifacts_url ? (
            <a
              href={task.artifacts_url}
              target="_blank"
              rel="noreferrer"
              className={styles.artifact}
            >
              <LinkIcon size={16} />
              {task.artifacts_url.replace(/^https?:\/\//, '')}
            </a>
          ) : null}

          <TaskComments
            taskId={task.id}
            workspaceId={workspaceId}
            readOnly={task.status === 'completed'}
          />
        </div>
      ) : null}

      {change.isError ? (
        <p className={styles.error} role="alert">
          {errorMessage(change.error)}
        </p>
      ) : null}
    </article>
  )
}
