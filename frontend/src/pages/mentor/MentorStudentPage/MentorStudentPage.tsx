import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  createTask,
  deleteTask,
  getMyWorkspaces,
  getTasks,
  getWorkspace,
  updateTask,
} from '../../../api/workspaces'
import type { Task, TaskInput, WorkspaceSummary } from '../../../api/workspaces'
import { Button } from '../../../components/ui/Button'
import { Modal } from '../../../components/ui/Modal'
import { Segmented } from '../../../components/ui/Segmented'
import { errorMessage } from '../../../lib/errorMessage'
import { MeetingReview } from '../../student/WorkspacePage/MeetingReview'
import { MeetingPlanner } from './MeetingPlanner'
import { MentorTaskItem } from './MentorTaskItem'
import { ReviewCard } from './ReviewCard'
import { TaskForm } from './TaskForm'
import styles from './MentorStudentPage.module.css'

function needsAttention(pair: WorkspaceSummary) {
  return Boolean(pair.tasks_on_review) || (pair.inactive_days ?? 0) >= 7
}

export function MentorStudentPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [formKey, setFormKey] = useState(0)
  const [editing, setEditing] = useState<Task | null>(null)
  const [removing, setRemoving] = useState<Task | null>(null)

  const list = useQuery({ queryKey: ['workspaces'], queryFn: getMyWorkspaces })
  const workspace = useQuery({ queryKey: ['workspace', id], queryFn: () => getWorkspace(id) })
  const tasks = useQuery({ queryKey: ['workspace', id, 'tasks'], queryFn: () => getTasks(id) })

  function refresh() {
    queryClient.invalidateQueries({ queryKey: ['workspace', id] })
    queryClient.invalidateQueries({ queryKey: ['workspaces'] })
  }

  const create = useMutation({
    mutationFn: (input: TaskInput) => createTask(id, input),
    onSuccess: () => {
      refresh()
      setFormKey((key) => key + 1)
    },
  })

  const edit = useMutation({
    mutationFn: ({ taskId, input }: { taskId: string; input: TaskInput }) =>
      updateTask(taskId, input),
    onSuccess: () => {
      refresh()
      setEditing(null)
    },
  })

  const remove = useMutation({
    mutationFn: (taskId: string) => deleteTask(taskId),
    onSuccess: () => {
      refresh()
      setRemoving(null)
    },
  })

  const tabs = (list.data?.items ?? []).map((pair) => ({
    value: pair.id,
    label: `${pair.partner.first_name}${needsAttention(pair) ? ' •' : ''}`,
  }))

  if (workspace.isPending) return <p className={styles.state}>Загружаем пространство ученика…</p>

  if (workspace.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось открыть пространство ученика.</p>
        <Button variant="secondary" to="/mentor/students">
          К списку учеников
        </Button>
      </div>
    )
  }

  const ws = workspace.data
  const allTasks = tasks.data ?? []
  const onReview = allTasks.filter((task) => task.status === 'on_review')

  return (
    <div className={styles.page}>
      {tabs.length > 1 ? (
        <div className={styles.tabs}>
          <Segmented
            options={tabs}
            value={id}
            onChange={(value) => navigate(`/mentor/students/${value}`)}
          />
        </div>
      ) : null}

      <div className={styles.columns}>
        <div className={styles.main}>
          <section className={styles.goal}>
            <div className={styles.goalText}>
              <span className={styles.muted}>
                {ws.partner.first_name} {ws.partner.last_name} · цель
              </span>
              <strong>{ws.goal ?? 'Цель уточните на первой встрече'}</strong>
            </div>
            <span className={styles.goalBar}>
              <span style={{ width: `${ws.progress_percent}%` }} />
            </span>
            <strong className={styles.goalPercent}>{ws.progress_percent}%</strong>
            <Button variant="secondary" size="sm" to="/mentor/chat">
              Чат
            </Button>
          </section>

          {onReview.map((task) => (
            <ReviewCard key={task.id} task={task} workspaceId={ws.id} />
          ))}

          <section className={styles.roadmap}>
            <div className={styles.roadmapHead}>
              <h2 className={styles.sectionTitle}>Дорожная карта</h2>
              <span className={styles.program}>
                {ws.service_package_title
                  ? `Программа «${ws.service_package_title}»`
                  : 'Открытый запрос'}
              </span>
            </div>

            {tasks.isPending ? <p className={styles.muted}>Загружаем задачи…</p> : null}

            {tasks.isSuccess && allTasks.length === 0 ? (
              <p className={styles.muted}>Задач пока нет. Добавьте первую в форме справа.</p>
            ) : null}

            {allTasks.map((task) => (
              <MentorTaskItem
                key={task.id}
                task={task}
                workspaceId={ws.id}
                onEdit={() => {
                  edit.reset()
                  setEditing(task)
                }}
                onDelete={() => {
                  remove.reset()
                  setRemoving(task)
                }}
              />
            ))}
          </section>
        </div>

        <aside className={styles.side}>
          <section className={styles.card}>
            <h3 className={styles.cardTitle}>Новая задача</h3>
            <TaskForm
              key={formKey}
              submitLabel="Добавить в дорожную карту"
              pending={create.isPending}
              error={create.isError ? create.error : undefined}
              onSubmit={(input) => create.mutate(input)}
            />
          </section>

          <MeetingPlanner workspace={ws} />

          <MeetingReview
            workspaceId={ws.id}
            ask={(date) => `Оцените встречу ${date}`}
            hint="Оценка видна только куратору и помогает ему видеть, как идут дела у пары."
            doneText="Куратор увидит её в карточке пары."
          />
        </aside>
      </div>

      {editing ? (
        <Modal title="Изменить задачу" onClose={() => setEditing(null)}>
          <TaskForm
            initial={editing}
            submitLabel="Сохранить"
            pending={edit.isPending}
            error={edit.isError ? edit.error : undefined}
            onSubmit={(input) => edit.mutate({ taskId: editing.id, input })}
            onCancel={() => setEditing(null)}
          />
        </Modal>
      ) : null}

      {removing ? (
        <Modal title="Удалить задачу?" onClose={() => setRemoving(null)}>
          <p className={styles.confirmText}>
            «{removing.title}» пропадёт из дорожной карты вместе с комментариями. Отменить это
            нельзя.
          </p>
          {remove.isError ? (
            <p className={styles.error} role="alert">
              {errorMessage(remove.error)}
            </p>
          ) : null}
          <div className={styles.formActions}>
            <Button variant="secondary" onClick={() => setRemoving(null)}>
              Отмена
            </Button>
            <Button
              variant="danger"
              disabled={remove.isPending}
              onClick={() => remove.mutate(removing.id)}
            >
              {remove.isPending ? 'Удаляем…' : 'Удалить'}
            </Button>
          </div>
        </Modal>
      ) : null}
    </div>
  )
}
