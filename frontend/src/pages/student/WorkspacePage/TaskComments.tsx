import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Send } from 'lucide-react'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { addTaskComment, getTaskComments } from '../../../api/workspaces'
import { Avatar } from '../../../components/ui/Avatar'
import { Input } from '../../../components/ui/Input'
import { formatDate } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import styles from './WorkspacePage.module.css'

type TaskCommentsProps = {
  taskId: string
  workspaceId: string
}

export function TaskComments({ taskId, workspaceId }: TaskCommentsProps) {
  const queryClient = useQueryClient()
  const [text, setText] = useState('')

  const comments = useQuery({
    queryKey: ['task', taskId, 'comments'],
    queryFn: () => getTaskComments(taskId),
  })

  const send = useMutation({
    mutationFn: () => addTaskComment(taskId, text.trim()),
    onSuccess: () => {
      setText('')
      queryClient.invalidateQueries({ queryKey: ['task', taskId, 'comments'] })
      queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId, 'tasks'] })
    },
  })

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (text.trim()) send.mutate()
  }

  return (
    <div className={styles.comments}>
      {comments.isPending ? <p className={styles.muted}>Загружаем комментарии…</p> : null}

      {comments.data?.map((comment) => {
        const name = `${comment.author.first_name} ${comment.author.last_name}`.trim()
        return (
          <div key={comment.id} className={styles.comment}>
            <Avatar name={name} size="s" />
            <div className={styles.commentBody}>
              <span className={styles.commentMeta}>
                {comment.author.first_name} · {formatDate(comment.created_at)}
              </span>
              <p className={styles.commentText}>{comment.text}</p>
            </div>
          </div>
        )
      })}

      <form className={styles.commentForm} onSubmit={handleSubmit}>
        <Input
          placeholder="Комментарий к задаче"
          aria-label="Комментарий к задаче"
          maxLength={2000}
          value={text}
          onChange={(e) => setText(e.target.value)}
        />
        <button
          type="submit"
          className={styles.sendButton}
          disabled={!text.trim() || send.isPending}
          aria-label="Отправить комментарий"
        >
          <Send size={18} />
        </button>
      </form>

      {send.isError ? (
        <p className={styles.error} role="alert">
          {errorMessage(send.error)}
        </p>
      ) : null}
    </div>
  )
}
