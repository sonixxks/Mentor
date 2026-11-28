import { useQuery } from '@tanstack/react-query'
import { CalendarDays, MessageCircle, Video } from 'lucide-react'
import { Link } from 'react-router-dom'
import { getMyWorkspaces, getTasks, getWorkspace } from '../../../api/workspaces'
import type { WorkspaceDetail } from '../../../api/workspaces'
import { Avatar } from '../../../components/ui/Avatar'
import { Button } from '../../../components/ui/Button'
import { formatMeetingTime, minutesBetween } from '../../../lib/dates'
import { plural } from '../../../lib/plural'
import { KnowledgeBase } from './KnowledgeBase'
import { MeetingReview } from './MeetingReview'
import { TaskItem } from './TaskItem'
import styles from './WorkspacePage.module.css'

function GoalCard({ workspace }: { workspace: WorkspaceDetail }) {
  const program = workspace.service_package_title
  const done = workspace.sessions_done ?? 0
  const sessions = workspace.sessions_total
    ? `встреч ${done} из ${workspace.sessions_total}`
    : `встреч проведено: ${done}`
  const title = workspace.goal
    ? `Цель: ${workspace.goal}`
    : program
      ? `Программа: ${program}`
      : 'Прогресс по задачам'

  return (
    <section className={styles.goal}>
      <div className={styles.goalMain}>
        <span className={styles.goalLabel}>{title}</span>
        <strong className={styles.goalPercent}>{workspace.progress_percent}%</strong>
      </div>
      <div className={styles.goalSide}>
        <span className={styles.goalBar}>
          <span style={{ width: `${workspace.progress_percent}%` }} />
        </span>
        <div className={styles.goalFacts}>
          <span>
            Выполнено задач: {workspace.tasks_completed ?? 0} из {workspace.tasks_total ?? 0}
          </span>
          <span>
            {program ?? 'Открытый запрос'} · {sessions}
          </span>
        </div>
      </div>
    </section>
  )
}

function PartnerCard({ workspace }: { workspace: WorkspaceDetail }) {
  const partner = workspace.partner
  const name = `${partner.first_name} ${partner.last_name}`.trim()

  return (
    <section className={styles.card}>
      <div className={styles.partner}>
        <Avatar name={name} size="m" />
        <div className={styles.partnerInfo}>
          <Link to={`/student/mentors/${partner.id}`} className={styles.partnerName}>
            {name}
          </Link>
          <span className={styles.muted}>наставник</span>
        </div>
        <Link
          to={`/student/chat/${workspace.id}`}
          className={styles.chatButton}
          aria-label="Написать наставнику"
        >
          <MessageCircle size={20} />
        </Link>
      </div>
    </section>
  )
}

function MeetingCard({ workspace }: { workspace: WorkspaceDetail }) {
  const meeting = workspace.next_meeting
  const link = meeting?.jitsi_link ?? workspace.jitsi_link
  const minutes = meeting ? minutesBetween(meeting.start_time, meeting.end_time) : null

  return (
    <section className={styles.card}>
      <span className={styles.muted}>Следующая встреча</span>

      {meeting ? (
        <>
          <strong className={styles.meetingTime}>{formatMeetingTime(meeting.start_time)}</strong>
          <p className={styles.secondary}>
            {[
              meeting.title,
              minutes ? plural(minutes, ['минута', 'минуты', 'минут']) : null,
              'видеозвонок',
            ]
              .filter(Boolean)
              .join(' · ')}
          </p>
        </>
      ) : (
        <p className={styles.secondary}>
          Встреча пока не назначена. Наставник назначит её в календаре.
        </p>
      )}

      <div className={styles.meetingActions}>
        {link ? (
          <a href={link} target="_blank" rel="noreferrer" className={styles.join}>
            <Video size={18} />
            Подключиться
          </a>
        ) : null}
        <Link to="/student/calendar" className={styles.iconButton} aria-label="Открыть календарь">
          <CalendarDays size={20} />
        </Link>
      </div>

      {link ? (
        <p className={styles.hint}>
          Jitsi Meet · {link.replace(/^https?:\/\//, '')} · ссылка постоянная для пары
        </p>
      ) : null}
    </section>
  )
}

export function WorkspacePage() {
  const list = useQuery({ queryKey: ['workspaces'], queryFn: getMyWorkspaces })
  const current =
    list.data?.items.find((item) => item.status !== 'completed') ?? list.data?.items[0]
  const id = current?.id ?? ''

  const workspace = useQuery({
    queryKey: ['workspace', id],
    queryFn: () => getWorkspace(id),
    enabled: Boolean(current),
  })
  const tasks = useQuery({
    queryKey: ['workspace', id, 'tasks'],
    queryFn: () => getTasks(id),
    enabled: Boolean(current),
  })

  if (list.isPending) return <p className={styles.state}>Загружаем пространство…</p>

  if (list.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось загрузить пространство.</p>
        <Button variant="secondary" onClick={() => list.refetch()}>
          Попробовать ещё раз
        </Button>
      </div>
    )
  }

  if (!current) {
    return (
      <div className={styles.state}>
        <h2 className={styles.stateTitle}>Пространства пока нет</h2>
        <p>
          Оно появится, когда наставник примет заявку и ты подтвердишь начало работы. Здесь будут
          цель, задачи, встречи и материалы.
        </p>
        <Button to="/student/requests">Мои заявки</Button>
      </div>
    )
  }

  if (workspace.isPending) return <p className={styles.state}>Загружаем пространство…</p>

  if (workspace.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось загрузить пространство.</p>
        <Button variant="secondary" onClick={() => workspace.refetch()}>
          Попробовать ещё раз
        </Button>
      </div>
    )
  }

  const ws = workspace.data

  return (
    <div className={styles.page}>
      <div className={styles.main}>
        <GoalCard workspace={ws} />

        <section className={styles.tasks}>
          <div className={styles.tasksHead}>
            <h2 className={styles.sectionTitle}>Трекер и задачи</h2>
            <span className={styles.muted}>Задачи ставит наставник</span>
          </div>

          {tasks.isPending ? <p className={styles.muted}>Загружаем задачи…</p> : null}

          {tasks.isError ? (
            <p className={styles.error}>Не удалось загрузить задачи. Обнови страницу.</p>
          ) : null}

          {tasks.data?.length === 0 ? (
            <p className={styles.muted}>
              Задач пока нет. Наставник добавит их после первой встречи.
            </p>
          ) : null}

          {tasks.data?.map((task) => (
            <TaskItem key={task.id} task={task} workspaceId={ws.id} />
          ))}
        </section>
      </div>

      <aside className={styles.side}>
        <PartnerCard workspace={ws} />
        <MeetingCard workspace={ws} />
        <MeetingReview workspaceId={ws.id} />
        <KnowledgeBase workspace={ws} />
      </aside>
    </div>
  )
}
