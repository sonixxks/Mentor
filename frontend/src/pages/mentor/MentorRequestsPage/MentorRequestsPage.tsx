import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { changeRequestStatus, getMyRequests } from '../../../api/requests'
import type { MentorRequest, RequestStatus } from '../../../api/requests'
import { Avatar } from '../../../components/ui/Avatar'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Modal } from '../../../components/ui/Modal'
import { Segmented } from '../../../components/ui/Segmented'
import { Tag } from '../../../components/ui/Tag'
import { Textarea } from '../../../components/ui/Textarea'
import { formatDate } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import { REQUEST_STATUS, STYLE_LABEL } from '../../../lib/labels'
import { GOALS } from '../../onboarding/goals'
import styles from './MentorRequestsPage.module.css'

type Tab = 'active' | 'accepted' | 'archive'

const TAB_STATUSES: Record<Tab, RequestStatus[]> = {
  active: ['pending', 'approved_by_mentor'],
  accepted: ['accepted'],
  archive: ['rejected_by_mentor', 'rejected_by_curator', 'cancelled'],
}

const EMPTY: Record<Tab, { title: string; text: string }> = {
  active: {
    title: 'Новых заявок нет',
    text: 'Когда ученик отправит заявку, она появится здесь. Чтобы заявок было больше, заполни профиль и добавь авторские программы.',
  },
  accepted: {
    title: 'Принятых заявок пока нет',
    text: 'Когда ученик подтвердит начало работы, заявка появится здесь.',
  },
  archive: {
    title: 'Архив пуст',
    text: 'Здесь появятся отклонённые и отменённые заявки.',
  },
}

const MENTOR_LABEL: Partial<Record<RequestStatus, string>> = {
  pending: 'Новая',
  approved_by_mentor: 'Ждёт ученика',
  rejected_by_mentor: 'Отклонена',
}

function studentFacts(request: MentorRequest) {
  const student = request.student
  if (!student) return ''
  const goal = GOALS.find((item) => item.value === student.main_goals?.[0])?.title
  const min = student.min_duration_months
  const max = student.max_duration_months
  const months = min && max ? (min === max ? `${min} мес.` : `${min}–${max} мес.`) : null

  return [
    goal,
    student.time_limit ? `${student.time_limit} ч в неделю` : null,
    months ? `срок ${months}` : null,
    student.communication_style
      ? `стиль: ${STYLE_LABEL[student.communication_style].toLowerCase()}`
      : null,
  ]
    .filter(Boolean)
    .join(' · ')
}

type RequestItemProps = {
  request: MentorRequest
  busy: boolean
  onApprove: () => void
  onReject: () => void
}

function RequestItem({ request, busy, onApprove, onReject }: RequestItemProps) {
  const name = request.student?.first_name ?? 'Ученик'
  const status = REQUEST_STATUS[request.status]
  const facts = studentFacts(request)
  const skills = request.student?.target_skills ?? []

  return (
    <article className={styles.item}>
      <div className={styles.head}>
        <Avatar name={name} size="m" />
        <div className={styles.info}>
          <span className={styles.name}>{name}</span>
          <span className={styles.meta}>
            {request.service_package_title ?? 'Открытый запрос'} · {formatDate(request.created_at)}
          </span>
        </div>
        {request.match ? <Badge variant="match">{request.match.percent}% совпадение</Badge> : null}
        <Badge variant={status.variant}>{MENTOR_LABEL[request.status] ?? status.label}</Badge>
      </div>

      {facts ? <p className={styles.facts}>{facts}</p> : null}

      {skills.length > 0 ? (
        <div className={styles.tags}>
          {skills.map((skill) => (
            <Tag key={skill}>{skill}</Tag>
          ))}
        </div>
      ) : null}

      {request.cover_letter ? <p className={styles.letter}>«{request.cover_letter}»</p> : null}

      {request.rejection_reason ? (
        <p className={styles.reason}>Причина отказа: {request.rejection_reason}</p>
      ) : null}

      {request.status === 'pending' ? (
        <div className={styles.actions}>
          <Button variant="secondary" size="sm" disabled={busy} onClick={onReject}>
            Отклонить
          </Button>
          <Button size="sm" disabled={busy} onClick={onApprove}>
            Принять
          </Button>
        </div>
      ) : null}

      {request.status === 'approved_by_mentor' ? (
        <p className={styles.waiting}>
          Заявка принята. Пара появится, когда {name} подтвердит начало работы.
        </p>
      ) : null}

      {request.status === 'accepted' ? (
        <div className={styles.actions}>
          <Button variant="secondary" size="sm" to="/mentor/students">
            Перейти к ученику
          </Button>
        </div>
      ) : null}
    </article>
  )
}

export function MentorRequestsPage() {
  const [tab, setTab] = useState<Tab>('active')
  const [rejecting, setRejecting] = useState<MentorRequest | null>(null)
  const [reason, setReason] = useState('')
  const queryClient = useQueryClient()

  const requests = useQuery({
    queryKey: ['requests', 'all'],
    queryFn: () => getMyRequests(),
  })

  const change = useMutation({
    mutationFn: ({ id, status, reason }: { id: string; status: RequestStatus; reason?: string }) =>
      changeRequestStatus(id, status, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['requests'] })
      setRejecting(null)
    },
  })

  const all = requests.data?.items ?? []
  const items = all.filter((item) => TAB_STATUSES[tab].includes(item.status))
  const newCount = all.filter((item) => item.status === 'pending').length

  const tabs = [
    { value: 'active', label: newCount > 0 ? `Активные · ${newCount} новых` : 'Активные' },
    { value: 'accepted', label: 'Принятые' },
    { value: 'archive', label: 'Архив' },
  ]

  function openReject(request: MentorRequest) {
    change.reset()
    setReason('')
    setRejecting(request)
  }

  return (
    <div className={styles.page}>
      <Segmented
        options={tabs}
        value={tab}
        onChange={(value) => {
          setTab(value as Tab)
          change.reset()
        }}
      />

      {change.isError && !rejecting ? (
        <p className={styles.error} role="alert">
          {errorMessage(change.error)}
        </p>
      ) : null}

      {requests.isPending ? <p className={styles.state}>Загружаем заявки…</p> : null}

      {requests.isError ? (
        <div className={styles.state}>
          <p>Не удалось загрузить заявки.</p>
          <Button variant="secondary" size="md" onClick={() => requests.refetch()}>
            Попробовать ещё раз
          </Button>
        </div>
      ) : null}

      {requests.isSuccess && items.length === 0 ? (
        <div className={styles.state}>
          <h2 className={styles.stateTitle}>{EMPTY[tab].title}</h2>
          <p>{EMPTY[tab].text}</p>
        </div>
      ) : null}

      {items.length > 0 ? (
        <div className={styles.list}>
          {items.map((request) => (
            <RequestItem
              key={request.id}
              request={request}
              busy={change.isPending}
              onApprove={() => change.mutate({ id: request.id, status: 'approved_by_mentor' })}
              onReject={() => openReject(request)}
            />
          ))}
        </div>
      ) : null}

      {rejecting ? (
        <Modal title="Отклонить заявку?" onClose={() => setRejecting(null)}>
          <p className={styles.confirmText}>
            {rejecting.student?.first_name ?? 'Ученик'} увидит, что заявка отклонена, и сможет
            выбрать другого наставника.
          </p>
          <Textarea
            label="Причина (необязательно)"
            rows={3}
            maxLength={300}
            placeholder="Например: сейчас не работаю с этим направлением"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
          />
          {change.isError ? (
            <p className={styles.error} role="alert">
              {errorMessage(change.error)}
            </p>
          ) : null}
          <div className={styles.confirmActions}>
            <Button variant="secondary" onClick={() => setRejecting(null)}>
              Отмена
            </Button>
            <Button
              variant="danger"
              disabled={change.isPending}
              onClick={() =>
                change.mutate({
                  id: rejecting.id,
                  status: 'rejected_by_mentor',
                  reason: reason.trim() || undefined,
                })
              }
            >
              {change.isPending ? 'Отклоняем…' : 'Отклонить'}
            </Button>
          </div>
        </Modal>
      ) : null}
    </div>
  )
}
