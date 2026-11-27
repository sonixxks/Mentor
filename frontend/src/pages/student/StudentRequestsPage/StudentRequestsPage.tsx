import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { changeRequestStatus, getMyRequests } from '../../../api/requests'
import type { MentorRequest } from '../../../api/requests'
import { Avatar } from '../../../components/ui/Avatar'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Modal } from '../../../components/ui/Modal'
import { Segmented } from '../../../components/ui/Segmented'
import { formatDate } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import { REQUEST_STATUS } from '../../../lib/labels'
import styles from './StudentRequestsPage.module.css'

type Group = 'active' | 'archive'

const MAX_ACTIVE = 3

function mentorName(request: MentorRequest) {
  return request.mentor ? `${request.mentor.first_name} ${request.mentor.last_name}` : 'Наставник'
}

type RequestItemProps = {
  request: MentorRequest
  busy: boolean
  onCancel: () => void
  onAccept: () => void
}

function RequestItem({ request, busy, onCancel, onAccept }: RequestItemProps) {
  const name = mentorName(request)
  const status = REQUEST_STATUS[request.status]

  return (
    <article className={styles.item}>
      <div className={styles.head}>
        <Avatar name={name} size="m" />
        <div className={styles.info}>
          <Link to={`/student/mentors/${request.mentor_id}`} className={styles.name}>
            {name}
          </Link>
          <span className={styles.meta}>
            {request.service_package_title ?? 'Открытый запрос'} · отправлена{' '}
            {formatDate(request.created_at)}
          </span>
        </div>
        <Badge variant={status.variant}>{status.label}</Badge>
      </div>

      {request.cover_letter ? <p className={styles.letter}>«{request.cover_letter}»</p> : null}

      {request.rejection_reason ? (
        <p className={styles.reason}>Причина: {request.rejection_reason}</p>
      ) : null}

      {request.status === 'pending' ? (
        <div className={styles.actions}>
          <span className={styles.hint}>Наставник ещё не ответил</span>
          <Button variant="secondary" size="sm" disabled={busy} onClick={onCancel}>
            Отменить заявку
          </Button>
        </div>
      ) : null}

      {request.status === 'approved_by_mentor' ? (
        <div className={clsx(styles.actions, styles.approved)}>
          <span>Наставник согласен работать с тобой. Подтверди, чтобы создать пару.</span>
          <Button size="sm" disabled={busy} onClick={onAccept}>
            Начать работу
          </Button>
        </div>
      ) : null}

      {request.status === 'accepted' ? (
        <div className={styles.actions}>
          <span className={styles.hint}>Общее пространство уже открыто</span>
          <Button variant="secondary" size="sm" to="/student/workspace">
            Перейти в пространство
          </Button>
        </div>
      ) : null}
    </article>
  )
}

export function StudentRequestsPage() {
  const [group, setGroup] = useState<Group>('active')
  const [accepting, setAccepting] = useState<MentorRequest | null>(null)
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const active = useQuery({
    queryKey: ['requests', 'active'],
    queryFn: () => getMyRequests('active'),
  })
  const archive = useQuery({
    queryKey: ['requests', 'archive'],
    queryFn: () => getMyRequests('archive'),
    enabled: group === 'archive',
  })

  const change = useMutation({
    mutationFn: ({ id, status }: { id: string; status: 'cancelled' | 'accepted' }) =>
      changeRequestStatus(id, status),
    onSuccess: (updated) => {
      queryClient.invalidateQueries({ queryKey: ['requests'] })
      if (updated.status === 'accepted') {
        queryClient.invalidateQueries({ queryKey: ['profile', 'me'] })
        navigate('/student/workspace')
      }
    },
  })

  const current = group === 'active' ? active : archive
  const activeCount = active.data?.total ?? 0
  const othersCount = activeCount - 1

  const tabs = [
    { value: 'active', label: activeCount > 0 ? `Активные · ${activeCount}` : 'Активные' },
    { value: 'archive', label: 'Архив' },
  ]

  return (
    <div className={styles.page}>
      <div className={styles.toolbar}>
        <Segmented
          options={tabs}
          value={group}
          onChange={(value) => {
            setGroup(value as Group)
            change.reset()
          }}
        />
        <span className={styles.counter}>
          Активных заявок: {activeCount} из {MAX_ACTIVE}
        </span>
      </div>

      {group === 'active' && activeCount >= MAX_ACTIVE ? (
        <p className={styles.notice}>
          Это максимум. Чтобы отправить новую заявку, отмени одну из текущих или дождись ответа.
        </p>
      ) : null}

      {change.isError && !accepting ? (
        <p className={styles.error} role="alert">
          {errorMessage(change.error)}
        </p>
      ) : null}

      {current.isPending ? <p className={styles.state}>Загружаем заявки…</p> : null}

      {current.isError ? (
        <div className={styles.state}>
          <p>Не удалось загрузить заявки.</p>
          <Button variant="secondary" size="md" onClick={() => current.refetch()}>
            Попробовать ещё раз
          </Button>
        </div>
      ) : null}

      {current.isSuccess && current.data.items.length === 0 ? (
        group === 'active' ? (
          <div className={styles.state}>
            <h2 className={styles.stateTitle}>Активных заявок нет</h2>
            <p>
              Выбери наставника в подборе или каталоге. Одновременно можно отправить до 3 заявок.
            </p>
            <Button to="/student">Перейти к подбору</Button>
          </div>
        ) : (
          <div className={styles.state}>
            <h2 className={styles.stateTitle}>Архив пуст</h2>
            <p>Здесь появятся принятые, отклонённые и отменённые заявки.</p>
          </div>
        )
      ) : null}

      {current.isSuccess && current.data.items.length > 0 ? (
        <div className={styles.list}>
          {current.data.items.map((request) => (
            <RequestItem
              key={request.id}
              request={request}
              busy={change.isPending}
              onCancel={() => change.mutate({ id: request.id, status: 'cancelled' })}
              onAccept={() => {
                change.reset()
                setAccepting(request)
              }}
            />
          ))}
        </div>
      ) : null}

      {accepting ? (
        <Modal title="Начать работу?" onClose={() => setAccepting(null)}>
          <p className={styles.confirmText}>
            Наставник: <b>{mentorName(accepting)}</b>. После подтверждения откроется общее
            пространство с трекером, задачами и чатом.
            {othersCount > 0
              ? ` Остальные активные заявки (${othersCount}) отменятся автоматически.`
              : ''}
          </p>
          {change.isError ? (
            <p className={styles.error} role="alert">
              {errorMessage(change.error)}
            </p>
          ) : null}
          <div className={styles.confirmActions}>
            <Button variant="secondary" onClick={() => setAccepting(null)}>
              Отмена
            </Button>
            <Button
              disabled={change.isPending}
              onClick={() => change.mutate({ id: accepting.id, status: 'accepted' })}
            >
              {change.isPending ? 'Создаём пару…' : 'Начать работу'}
            </Button>
          </div>
        </Modal>
      ) : null}
    </div>
  )
}
