import type {
  Meeting,
  Task,
  TaskComment,
  WorkspaceFile,
  WorkspaceStatus,
} from '../../api/workspaces'
import { MENTORS } from './mentors'
import type { MockRequest } from './requests'

export type MockWorkspace = {
  id: string
  student_id: string
  mentor_id: string
  service_package_id?: string
  status: WorkspaceStatus
  started_at: string
}

export type MockTask = Task & { workspace_id: string }

export type MockMeeting = Omit<Meeting, 'jitsi_link' | 'my_review_rating'> & {
  reviews: Record<string, number>
}

export type MockFile = WorkspaceFile & { workspace_id: string }

type Store = {
  workspaces: MockWorkspace[]
  tasks: MockTask[]
  comments: TaskComment[]
  meetings: MockMeeting[]
  files: MockFile[]
}

const KEY = 'mock-workspaces'
const DAY = 24 * 60 * 60 * 1000

const empty = (): Store => ({ workspaces: [], tasks: [], comments: [], meetings: [], files: [] })

function load(): Store {
  try {
    const saved = localStorage.getItem(KEY)
    return saved ? JSON.parse(saved) : empty()
  } catch {
    return empty()
  }
}

export const store: Store = load()

export function saveStore() {
  localStorage.setItem(KEY, JSON.stringify(store))
}

function daysFromNow(days: number, hourInEkb = 12) {
  const date = new Date(Date.now() + days * DAY)
  date.setUTCHours(hourInEkb - 5, 0, 0, 0)
  return date.toISOString()
}

export function createWorkspace(request: MockRequest) {
  const id = `ws-${request.id}`
  const mentor = MENTORS.find((m) => m.id === request.mentor_id)

  store.workspaces.push({
    id,
    student_id: request.student_id,
    mentor_id: request.mentor_id,
    service_package_id: request.service_package_id,
    status: 'active',
    started_at: daysFromNow(-21),
  })

  const task = (n: number, fields: Omit<MockTask, 'id' | 'workspace_id' | 'created_at'>) => ({
    id: `${id}-t${n}`,
    workspace_id: id,
    created_at: daysFromNow(-20),
    ...fields,
  })

  store.tasks.push(
    task(1, {
      title: 'Настроить виртуальное окружение',
      status: 'completed',
      deadline_at: daysFromNow(-16),
      completed_at: daysFromNow(-17),
    }),
    task(2, {
      title: 'Написать тест-кейсы для проверки алгоритма',
      status: 'completed',
      deadline_at: daysFromNow(-9),
      completed_at: daysFromNow(-10),
      artifacts_url: 'https://github.com/ekaterina/test-cases',
    }),
    task(3, {
      title: 'Собрать и описать датасет',
      description: 'Источник, признаки, баланс классов. Оформить в README.',
      status: 'on_review',
      deadline_at: daysFromNow(-1),
      artifacts_url: 'https://github.com/ekaterina/dataset',
    }),
    task(4, {
      title: 'Протестировать модель на новых данных',
      description: 'Прогнать модель на свежей выборке и сравнить метрики с базовой версией.',
      status: 'in_progress',
      deadline_at: daysFromNow(5),
    }),
    task(5, {
      title: 'Подготовить отчёт по метрикам',
      status: 'todo',
      deadline_at: daysFromNow(12),
    }),
  )

  if (mentor) {
    store.comments.push({
      id: `${id}-c1`,
      task_id: `${id}-t4`,
      author: { id: mentor.id, first_name: mentor.first_name, last_name: mentor.last_name },
      text: 'Возьми датасет из базы знаний, там уже размечено 2000 примеров.',
      created_at: daysFromNow(-2),
    })
  }

  store.meetings.push(
    {
      id: `${id}-m1`,
      workspace_id: id,
      title: 'Знакомство и цели',
      start_time: daysFromNow(-7, 18),
      end_time: daysFromNow(-7, 19),
      reviews: {},
    },
    {
      id: `${id}-m2`,
      workspace_id: id,
      title: 'Разбор тест-кейсов',
      start_time: daysFromNow(2, 18),
      end_time: daysFromNow(2, 19),
      reviews: {},
    },
  )

  store.files.push(
    {
      id: `${id}-f1`,
      workspace_id: id,
      type: 'file',
      file_name: 'План работы.docx',
      file_url: 'https://example.com/plan.docx',
      file_size_kb: 248,
      uploaded_by: request.mentor_id,
      created_at: daysFromNow(-20),
    },
    {
      id: `${id}-f2`,
      workspace_id: id,
      type: 'link',
      file_name: 'Метрики качества моделей',
      file_url: 'https://scikit-learn.org/stable/modules/model_evaluation.html',
      file_size_kb: null,
      uploaded_by: request.mentor_id,
      created_at: daysFromNow(-5),
    },
  )

  saveStore()
}
