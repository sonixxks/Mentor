import { clsx } from 'clsx'
import { ChevronDown, Link as LinkIcon, Pencil, Trash2 } from 'lucide-react'
import { useState } from 'react'
import type { Task } from '../../../api/workspaces'
import { Badge } from '../../../components/ui/Badge'
import { formatDate, isPast } from '../../../lib/dates'
import { TASK_STATUS } from '../../../lib/labels'
import { plural } from '../../../lib/plural'
import { TaskComments } from '../../student/WorkspacePage/TaskComments'
import styles from './MentorStudentPage.module.css'

type MentorTaskItemProps = {
  task: Task
  workspaceId: string
  onEdit: () => void
  onDelete: () => void
}

function taskMeta(task: Task) {
  const parts: string[] = []
  if (task.status === 'completed' && task.completed_at) {
    parts.push(`выполнено ${formatDate(task.completed_at)}`)
  } else if (task.status === 'on_review') {
    parts.push('ждёт вашей проверки')
  } else {
    parts.push(`до ${formatDate(task.deadline_at)}`)
  }
  if (task.comments_count) {
    parts.push(plural(task.comments_count, ['комментарий', 'комментария', 'комментариев']))
  }
  return parts.join(' · ')
}

export function MentorTaskItem({ task, workspaceId, onEdit, onDelete }: MentorTaskItemProps) {
  const [open, setOpen] = useState(false)
  const status = TASK_STATUS[task.status]
  const editable = task.status !== 'completed' && task.status !== 'on_review'
  const overdue =
    isPast(task.deadline_at) && task.status !== 'completed' && task.status !== 'on_review'

  function toggle() {
    setOpen((value) => !value)
  }

  return (
    <article className={styles.task}>
      <div className={styles.taskHead}>
        <button type="button" className={styles.taskTitle} aria-expanded={open} onClick={toggle}>
          <span className={styles.taskName}>{task.title}</span>
          <span className={clsx(styles.taskMeta, overdue ? styles.overdue : '')}>
            {overdue ? 'просрочено · ' : ''}
            {taskMeta(task)}
          </span>
        </button>

        <Badge variant={status.variant}>{status.label}</Badge>

        <div className={styles.taskActions}>
          {editable ? (
            <>
              <button
                type="button"
                className={styles.iconButton}
                aria-label={`Изменить задачу «${task.title}»`}
                onClick={onEdit}
              >
                <Pencil size={18} />
              </button>
              <button
                type="button"
                className={styles.iconButton}
                aria-label={`Удалить задачу «${task.title}»`}
                onClick={onDelete}
              >
                <Trash2 size={18} />
              </button>
            </>
          ) : null}
        </div>
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

          {task.artifacts_url ? (
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

          {task.mentor_comment ? (
            <p className={styles.mentorComment}>
              <b>Ваш отзыв:</b> {task.mentor_comment}
            </p>
          ) : null}

          <TaskComments taskId={task.id} workspaceId={workspaceId} />
        </div>
      ) : null}
    </article>
  )
}
