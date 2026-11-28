import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Link as LinkIcon } from 'lucide-react'
import { useState } from 'react'
import { changeTaskStatus } from '../../../api/workspaces'
import type { Task } from '../../../api/workspaces'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Textarea } from '../../../components/ui/Textarea'
import { formatDate } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import styles from './MentorStudentPage.module.css'

type ReviewCardProps = {
  task: Task
  workspaceId: string
}

export function ReviewCard({ task, workspaceId }: ReviewCardProps) {
  const queryClient = useQueryClient()
  const [returning, setReturning] = useState(false)
  const [comment, setComment] = useState('')

  const decide = useMutation({
    mutationFn: (status: 'completed' | 'needs_revision') =>
      changeTaskStatus(task.id, { status, mentor_comment: comment.trim() || undefined }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId] })
      queryClient.invalidateQueries({ queryKey: ['workspaces'] })
    },
  })

  return (
    <section className={styles.review}>
      <div className={styles.reviewHead}>
        <Badge variant="review">Ждёт вашей проверки</Badge>
        <span className={styles.muted}>
          отправлено {formatDate(task.updated_at ?? task.created_at)}
        </span>
      </div>

      <h3 className={styles.reviewTitle}>{task.title}</h3>

      {task.artifacts_url ? (
        <a href={task.artifacts_url} target="_blank" rel="noreferrer" className={styles.artifact}>
          <LinkIcon size={16} />
          {task.artifacts_url.replace(/^https?:\/\//, '')}
        </a>
      ) : (
        <p className={styles.muted}>Ученик отправил без ссылки — результат обсудите на встрече.</p>
      )}

      {returning ? (
        <>
          <Textarea
            label="Что доработать"
            rows={3}
            maxLength={2000}
            placeholder="Ученик увидит этот текст над задачей"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
          />
          <div className={styles.actions}>
            <Button variant="secondary" onClick={() => setReturning(false)}>
              Отмена
            </Button>
            <Button
              disabled={!comment.trim() || decide.isPending}
              onClick={() => decide.mutate('needs_revision')}
            >
              {decide.isPending ? 'Отправляем…' : 'Вернуть на доработку'}
            </Button>
          </div>
        </>
      ) : (
        <div className={styles.actions}>
          <Button disabled={decide.isPending} onClick={() => decide.mutate('completed')}>
            Принять
          </Button>
          <Button variant="secondary" onClick={() => setReturning(true)}>
            Вернуть на доработку
          </Button>
        </div>
      )}

      {decide.isError ? (
        <p className={styles.error} role="alert">
          {errorMessage(decide.error)}
        </p>
      ) : null}
    </section>
  )
}
