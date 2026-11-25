import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { Star } from 'lucide-react'
import { useState } from 'react'
import { getMeetings, reviewMeeting } from '../../../api/workspaces'
import { Button } from '../../../components/ui/Button'
import { Textarea } from '../../../components/ui/Textarea'
import { formatDate, isPast } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import styles from './WorkspacePage.module.css'

const STARS = [1, 2, 3, 4, 5]

export function MeetingReview({ workspaceId }: { workspaceId: string }) {
  const queryClient = useQueryClient()
  const [rating, setRating] = useState(0)
  const [hover, setHover] = useState(0)
  const [comment, setComment] = useState('')

  const meetings = useQuery({
    queryKey: ['workspace', workspaceId, 'meetings'],
    queryFn: () => getMeetings(workspaceId),
  })

  const send = useMutation({
    mutationFn: (meetingId: string) =>
      reviewMeeting(meetingId, rating, comment.trim() || undefined),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId, 'meetings'] }),
  })

  if (send.isSuccess) {
    return (
      <section className={clsx(styles.card, styles.reviewCard)}>
        <h3 className={styles.cardTitle}>Спасибо за оценку!</h3>
        <p className={styles.secondary}>Она появится в отзывах наставника.</p>
      </section>
    )
  }

  const target = [...(meetings.data ?? [])]
    .reverse()
    .find((meeting) => isPast(meeting.start_time) && meeting.my_review_rating == null)

  if (!target) return null

  const shown = hover || rating

  return (
    <section className={clsx(styles.card, styles.reviewCard)}>
      <h3 className={styles.cardTitle}>Как прошла встреча {formatDate(target.start_time)}?</h3>

      <div
        className={styles.stars}
        role="radiogroup"
        aria-label="Оценка встречи"
        onMouseLeave={() => setHover(0)}
      >
        {STARS.map((value) => (
          <button
            key={value}
            type="button"
            role="radio"
            aria-checked={rating === value}
            aria-label={`${value} из 5`}
            className={styles.starButton}
            onMouseEnter={() => setHover(value)}
            onClick={() => setRating(value)}
          >
            <Star
              size={28}
              fill={value <= shown ? 'currentColor' : 'none'}
              className={clsx(styles.star, value <= shown ? styles.starOn : '')}
            />
          </button>
        ))}
      </div>

      {rating > 0 ? (
        <>
          <Textarea
            rows={3}
            maxLength={2000}
            placeholder="Пара слов о встрече, необязательно"
            value={comment}
            onChange={(e) => setComment(e.target.value)}
          />
          <Button size="md" disabled={send.isPending} onClick={() => send.mutate(target.id)}>
            {send.isPending ? 'Отправляем…' : 'Оценить'}
          </Button>
        </>
      ) : null}

      <p className={styles.hint}>
        Оценка с твоим именем и датой встречи появится в отзывах наставника.
      </p>

      {send.isError ? (
        <p className={styles.error} role="alert">
          {errorMessage(send.error)}
        </p>
      ) : null}
    </section>
  )
}
