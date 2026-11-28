import { useQuery } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { getCalendar } from '../../../api/calendar'
import { getMyMentorStats, getMyReviews } from '../../../api/mentors'
import { getMyRequests } from '../../../api/requests'
import { getMyWorkspaces } from '../../../api/workspaces'
import type { Meeting, WorkspaceSummary } from '../../../api/workspaces'
import { RescheduleModal } from '../../../components/RescheduleModal'
import { Avatar } from '../../../components/ui/Avatar'
import { Button } from '../../../components/ui/Button'
import { dayLabel, formatDate, formatTime, todayEkb } from '../../../lib/dates'
import { reviewsLabel } from '../../../lib/labels'
import { plural } from '../../../lib/plural'
import styles from './MentorDashboard.module.css'

const UPCOMING_DAYS = 14
const INACTIVE_DAYS = 7

function studentStatus(workspace: WorkspaceSummary) {
  const inactive = workspace.inactive_days ?? 0
  if (inactive >= INACTIVE_DAYS) {
    return {
      text: `нет активности ${plural(inactive, ['день', 'дня', 'дней'])}`,
      className: styles.statusWarning,
      warning: true,
    }
  }
  if (workspace.tasks_on_review) {
    return {
      text: `${plural(workspace.tasks_on_review, ['задача', 'задачи', 'задач'])} на проверке`,
      className: styles.statusAccent,
      warning: false,
    }
  }
  return { text: 'по плану', className: '', warning: false }
}

export function MentorDashboard() {
  const [moving, setMoving] = useState<Meeting | null>(null)
  const from = todayEkb()
  const to = todayEkb(UPCOMING_DAYS)

  const stats = useQuery({ queryKey: ['mentor-stats'], queryFn: getMyMentorStats })
  const workspaces = useQuery({ queryKey: ['workspaces'], queryFn: getMyWorkspaces })
  const calendar = useQuery({
    queryKey: ['calendar', from, to],
    queryFn: () => getCalendar(from, to),
  })
  const requests = useQuery({
    queryKey: ['requests', 'active'],
    queryFn: () => getMyRequests('active'),
  })
  const reviews = useQuery({ queryKey: ['my-reviews'], queryFn: () => getMyReviews() })

  const pairs = workspaces.data?.items ?? []
  const partnerName = (workspaceId: string) =>
    pairs.find((pair) => pair.id === workspaceId)?.partner.first_name ?? 'Ученик'

  const upcoming = (calendar.data?.meetings ?? [])
    .filter((meeting) => Date.parse(meeting.end_time ?? meeting.start_time) > Date.now())
    .slice(0, 3)
  const newRequests = (requests.data?.items ?? []).filter((item) => item.status === 'pending')

  return (
    <div className={styles.page}>
      <section className={styles.tiles}>
        <div className={clsx(styles.tile, styles.tileAccent)}>
          <span className={styles.tileLabel}>Рейтинг</span>
          <div className={styles.tileValue}>
            <strong>
              {stats.data?.rating_avg != null
                ? stats.data.rating_avg.toFixed(1).replace('.', ',')
                : '—'}
            </strong>
            <span>
              {stats.data?.reviews_count
                ? `из 5 · ${reviewsLabel(stats.data.reviews_count)}`
                : 'пока нет отзывов'}
            </span>
          </div>
        </div>

        <div className={styles.tile}>
          <span className={styles.tileLabel}>Активные ученики</span>
          <div className={styles.tileValue}>
            <strong>{stats.data?.active_students ?? '—'}</strong>
            {stats.data ? (
              <span>из {plural(stats.data.max_students, ['места', 'мест', 'мест'])}</span>
            ) : null}
          </div>
        </div>

        <div className={styles.tile}>
          <span className={styles.tileLabel}>Завершено пар</span>
          <div className={styles.tileValue}>
            <strong>{stats.data?.completed_pairs ?? '—'}</strong>
            <span>за всё время</span>
          </div>
        </div>
      </section>

      <div className={styles.columns}>
        <div className={styles.column}>
          <section className={styles.card}>
            <div className={styles.cardHead}>
              <h2 className={styles.cardTitle}>Предстоящие встречи</h2>
              <Link to="/mentor/calendar" className={styles.cardLink}>
                Расписание
              </Link>
            </div>

            {calendar.isPending ? <p className={styles.muted}>Загружаем встречи…</p> : null}

            {calendar.isSuccess && upcoming.length === 0 ? (
              <p className={styles.muted}>
                На ближайшие две недели встреч нет. Назначить встречу можно на странице ученика.
              </p>
            ) : null}

            {upcoming.map((meeting) => {
              const name = partnerName(meeting.workspace_id)
              const day = dayLabel(meeting.start_time)
              return (
                <div key={meeting.id} className={styles.meeting}>
                  <div className={styles.meetingTime}>
                    <strong>{formatTime(meeting.start_time)}</strong>
                    <span>{day}</span>
                  </div>
                  <Avatar name={name} size="m" />
                  <div className={styles.info}>
                    <span className={styles.name}>{name}</span>
                    <span className={styles.muted}>{meeting.title}</span>
                  </div>
                  {day === 'сегодня' ? (
                    <a
                      href={meeting.jitsi_link}
                      target="_blank"
                      rel="noreferrer"
                      className={styles.join}
                    >
                      Подключиться
                    </a>
                  ) : (
                    <Button variant="secondary" size="sm" onClick={() => setMoving(meeting)}>
                      Перенести
                    </Button>
                  )}
                </div>
              )
            })}
          </section>

          <section className={styles.card}>
            <div className={styles.cardHead}>
              <h2 className={styles.cardTitle}>Мои ученики</h2>
              <Link to="/mentor/students" className={styles.cardLink}>
                Все
              </Link>
            </div>

            {workspaces.isPending ? <p className={styles.muted}>Загружаем учеников…</p> : null}

            {workspaces.isSuccess && pairs.length === 0 ? (
              <p className={styles.muted}>
                Учеников пока нет. Они появятся, когда вы примете заявку и ученик подтвердит начало
                работы.
              </p>
            ) : null}

            {pairs.map((pair) => {
              const status = studentStatus(pair)
              return (
                <Link key={pair.id} to={`/mentor/students/${pair.id}`} className={styles.student}>
                  <Avatar name={pair.partner.first_name} size="m" />
                  <span className={styles.name}>{pair.partner.first_name}</span>
                  <span className={clsx(styles.bar, status.warning ? styles.barWarning : '')}>
                    <span style={{ width: `${pair.progress_percent}%` }} />
                  </span>
                  <span className={styles.percent}>{pair.progress_percent}%</span>
                  <span className={clsx(styles.status, status.className)}>{status.text}</span>
                </Link>
              )
            })}
          </section>
        </div>

        <div className={styles.column}>
          <section className={clsx(styles.card, styles.cardAccent)}>
            <div className={styles.cardHead}>
              <h2 className={styles.cardTitle}>
                Новые заявки{newRequests.length > 0 ? ` · ${newRequests.length}` : ''}
              </h2>
              <Link to="/mentor/requests" className={styles.cardLink}>
                Открыть
              </Link>
            </div>

            {requests.isSuccess && newRequests.length === 0 ? (
              <p className={styles.muted}>Новых заявок нет.</p>
            ) : null}

            {newRequests.slice(0, 3).map((request) => {
              const name = request.student?.first_name ?? 'Ученик'
              const about =
                request.student?.target_skills?.join(', ') ||
                request.service_package_title ||
                'Открытый запрос'
              return (
                <Link key={request.id} to="/mentor/requests" className={styles.request}>
                  <Avatar name={name} size="m" />
                  <div className={styles.info}>
                    <span className={styles.name}>{name}</span>
                    <span className={styles.muted}>{about}</span>
                  </div>
                  {request.match ? (
                    <span className={styles.match}>{request.match.percent}%</span>
                  ) : null}
                </Link>
              )
            })}
          </section>

          <section className={styles.card}>
            <h2 className={styles.cardTitle}>Последние отзывы</h2>

            {reviews.isSuccess && reviews.data.items.length === 0 ? (
              <p className={styles.muted}>Отзывов пока нет. Они появятся после первых встреч.</p>
            ) : null}

            {reviews.data?.items.map((review) => (
              <article key={review.id} className={styles.review}>
                <span className={styles.stars} aria-label={`${review.rating} из 5`}>
                  {'★'.repeat(review.rating)}
                  {'☆'.repeat(5 - review.rating)}
                </span>
                {review.comment ? <p className={styles.reviewText}>«{review.comment}»</p> : null}
                <span className={styles.muted}>
                  {review.author_first_name}
                  {review.meeting_date ? ` · встреча ${formatDate(review.meeting_date)}` : ''}
                </span>
              </article>
            ))}
          </section>
        </div>
      </div>

      {moving ? (
        <RescheduleModal
          meeting={moving}
          partnerName={partnerName(moving.workspace_id)}
          onClose={() => setMoving(null)}
        />
      ) : null}
    </div>
  )
}
