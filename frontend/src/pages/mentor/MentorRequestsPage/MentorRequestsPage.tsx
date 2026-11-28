import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { useState } from 'react'
import { getMyMentorStats } from '../../../api/mentors'
import { changeRequestStatus, getMyRequests } from '../../../api/requests'
import type { MentorRequest, RequestStatus } from '../../../api/requests'
import { Avatar } from '../../../components/ui/Avatar'
import { Badge } from '../../../components/ui/Badge'
import { Button } from '../../../components/ui/Button'
import { Modal } from '../../../components/ui/Modal'
import { Segmented } from '../../../components/ui/Segmented'
import { Tag } from '../../../components/ui/Tag'
import { Textarea } from '../../../components/ui/Textarea'
import { useMyProfile } from '../../../hooks/useMyProfile'
import { formatDate, sentLabel } from '../../../lib/dates'
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

const EMPTY: Record<Tab, string> = {
  active:
    'Новых заявок нет. Чтобы их было больше, заполните профиль и добавьте авторские программы.',
  accepted: 'Принятых заявок пока нет.',
  archive: 'Здесь появятся отклонённые и отменённые заявки.',
}

const MENTOR_LABEL: Partial<Record<RequestStatus, string>> = {
  pending: 'Новая',
  approved_by_mentor: 'Ждёт ученика',
  rejected_by_mentor: 'Отклонена',
}

function studentName(request: MentorRequest) {
  return request.student?.first_name ?? 'Ученик'
}

function goalTitle(request: MentorRequest) {
  return GOALS.find((item) => item.value === request.student?.main_goals?.[0])?.title
}

function timeAndTerm(request: MentorRequest) {
  const student = request.student
  const min = student?.min_duration_months
  const max = student?.max_duration_months
  const months = min && max ? (min === max ? `${min} мес.` : `${min}–${max} мес.`) : null
  return [student?.time_limit ? `${student.time_limit} ч в неделю` : null, months]
    .filter(Boolean)
    .join(' · ')
}

type DetailProps = {
  request: MentorRequest
  mySkills: Set<string>
  busy: boolean
  error: unknown
  onApprove: () => void
  onReject: () => void
}

function RequestDetail({ request, mySkills, busy, error, onApprove, onReject }: DetailProps) {
  const name = studentName(request)
  const style = request.student?.communication_style
  const skills = request.student?.target_skills ?? []

  const facts = [
    { label: 'Цель', value: goalTitle(request) },
    { label: 'Время и срок', value: timeAndTerm(request) },
    { label: 'Стиль общения', value: style ? STYLE_LABEL[style] : undefined },
  ]

  return (
    <section className={styles.detail}>
      <div className={styles.detailHead}>
        <Avatar name={name} size="l" />
        <div className={styles.detailInfo}>
          <h2 className={styles.detailName}>{name}</h2>
          <span className={styles.secondary}>
            {request.service_package_title
              ? `Программа «${request.service_package_title}»`
              : 'Открытый запрос'}{' '}
            · заявка от {formatDate(request.created_at)}
          </span>
        </div>
        {request.match ? (
          <div className={styles.matchValue}>
            <strong>{request.match.percent}%</strong>
            <span>совместимость</span>
          </div>
        ) : null}
      </div>

      {request.cover_letter ? (
        <p className={styles.letter}>«{request.cover_letter}»</p>
      ) : (
        <p className={styles.muted}>Ученик не написал сопроводительное письмо.</p>
      )}

      <div className={styles.facts}>
        {facts.map((fact) => (
          <div key={fact.label} className={styles.fact}>
            <span className={styles.factLabel}>{fact.label}</span>
            <strong>{fact.value || '—'}</strong>
          </div>
        ))}
      </div>

      {skills.length > 0 ? (
        <div className={styles.interests}>
          <span className={styles.factLabel}>Интересы</span>
          <div className={styles.tags}>
            {skills.map((skill) => (
              <Tag key={skill} selected={mySkills.has(skill)}>
                {mySkills.has(skill) ? `${skill} · совпадает` : skill}
              </Tag>
            ))}
          </div>
        </div>
      ) : null}

      {request.rejection_reason ? (
        <p className={styles.reason}>Причина отказа: {request.rejection_reason}</p>
      ) : null}

      {error ? (
        <p className={styles.error} role="alert">
          {errorMessage(error)}
        </p>
      ) : null}

      <div className={styles.detailFooter}>
        {request.status === 'pending' ? (
          <>
            <Button disabled={busy} onClick={onApprove}>
              Принять заявку
            </Button>
            <Button variant="danger" disabled={busy} onClick={onReject}>
              Отклонить
            </Button>
            <span className={styles.muted}>
              После принятия {name} подтвердит выбор, и пространство пары создастся автоматически.
            </span>
          </>
        ) : null}

        {request.status === 'approved_by_mentor' ? (
          <span className={styles.secondary}>
            Вы приняли заявку. Пара появится, когда {name} подтвердит начало работы.
          </span>
        ) : null}

        {request.status === 'accepted' && request.workspace_id ? (
          <Button variant="secondary" to={`/mentor/students/${request.workspace_id}`}>
            Открыть пространство ученика
          </Button>
        ) : null}
      </div>
    </section>
  )
}

export function MentorRequestsPage() {
  const [tab, setTab] = useState<Tab>('active')
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [rejecting, setRejecting] = useState<MentorRequest | null>(null)
  const [reason, setReason] = useState('')
  const queryClient = useQueryClient()
  const profile = useMyProfile()

  const requests = useQuery({
    queryKey: ['requests', 'all'],
    queryFn: () => getMyRequests(),
  })
  const stats = useQuery({ queryKey: ['mentor-stats'], queryFn: getMyMentorStats })

  const change = useMutation({
    mutationFn: ({ id, status, reason }: { id: string; status: RequestStatus; reason?: string }) =>
      changeRequestStatus(id, status, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['requests'] })
      queryClient.invalidateQueries({ queryKey: ['mentor-stats'] })
      setRejecting(null)
    },
  })

  const all = requests.data?.items ?? []
  const items = all.filter((item) => TAB_STATUSES[tab].includes(item.status))
  const selected = items.find((item) => item.id === selectedId) ?? items[0]
  const newCount = all.filter((item) => item.status === 'pending').length
  const mySkills = new Set(profile.data?.current_skills?.map((item) => item.skill))

  const tabs = [
    { value: 'active', label: newCount > 0 ? `Новые · ${newCount}` : 'Новые' },
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
      <div className={styles.toolbar}>
        <Segmented
          options={tabs}
          value={tab}
          onChange={(value) => {
            setTab(value as Tab)
            setSelectedId(null)
            change.reset()
          }}
        />
        {stats.data ? (
          <span className={styles.slots}>
            Свободно мест:{' '}
            <b>
              {stats.data.free_slots} из {stats.data.max_students}
            </b>
          </span>
        ) : null}
      </div>

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
        <p className={styles.state}>{EMPTY[tab]}</p>
      ) : null}

      {selected ? (
        <div className={styles.layout}>
          <div className={styles.list}>
            {items.map((request) => {
              const name = studentName(request)
              const status = REQUEST_STATUS[request.status]
              return (
                <button
                  key={request.id}
                  type="button"
                  className={clsx(styles.item, request.id === selected.id ? styles.itemActive : '')}
                  aria-pressed={request.id === selected.id}
                  onClick={() => {
                    change.reset()
                    setSelectedId(request.id)
                  }}
                >
                  <span className={styles.itemHead}>
                    <Avatar name={name} size="m" />
                    <span className={styles.itemInfo}>
                      <span className={styles.itemName}>{name}</span>
                      <span className={styles.muted}>{sentLabel(request.created_at)}</span>
                    </span>
                    {request.match ? (
                      <span className={styles.match}>{request.match.percent}%</span>
                    ) : null}
                  </span>
                  <span className={styles.itemSkills}>
                    {request.student?.target_skills?.join(' · ') ||
                      request.service_package_title ||
                      'Открытый запрос'}
                  </span>
                  {request.status !== 'pending' ? (
                    <Badge variant={status.variant}>
                      {MENTOR_LABEL[request.status] ?? status.label}
                    </Badge>
                  ) : null}
                </button>
              )
            })}
          </div>

          <RequestDetail
            request={selected}
            mySkills={mySkills}
            busy={change.isPending}
            error={change.isError && !rejecting ? change.error : null}
            onApprove={() => change.mutate({ id: selected.id, status: 'approved_by_mentor' })}
            onReject={() => openReject(selected)}
          />
        </div>
      ) : null}

      {rejecting ? (
        <Modal title="Отклонить заявку?" onClose={() => setRejecting(null)}>
          <p className={styles.confirmText}>
            {studentName(rejecting)} увидит, что заявка отклонена, и сможет выбрать другого
            наставника.
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
