import { useQuery } from '@tanstack/react-query'
import { Navigate } from 'react-router-dom'
import { getMyWorkspaces } from '../../../api/workspaces'
import { Button } from '../../../components/ui/Button'
import styles from './MentorStudentPage.module.css'

export function MentorStudentsIndex() {
  const list = useQuery({ queryKey: ['workspaces'], queryFn: getMyWorkspaces })

  if (list.isPending) return <p className={styles.state}>Загружаем учеников…</p>

  if (list.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось загрузить учеников.</p>
        <Button variant="secondary" onClick={() => list.refetch()}>
          Попробовать ещё раз
        </Button>
      </div>
    )
  }

  const first = list.data.items.find((pair) => pair.status !== 'completed') ?? list.data.items[0]

  if (!first) {
    return (
      <div className={styles.state}>
        <h2 className={styles.stateTitle}>Учеников пока нет</h2>
        <p>
          Они появятся, когда вы примете заявку и ученик подтвердит начало работы. Новые заявки — во
          «Входящих».
        </p>
        <Button to="/mentor/requests">Входящие заявки</Button>
      </div>
    )
  }

  return <Navigate to={`/mentor/students/${first.id}`} replace />
}
