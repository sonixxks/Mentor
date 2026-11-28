import { useQuery } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { useParams } from 'react-router-dom'
import { getMyWorkspaces } from '../../../api/workspaces'
import { Button } from '../../../components/ui/Button'
import { useMe } from '../../../hooks/useMe'
import { useSession } from '../../../hooks/useSession'
import { Conversation } from './Conversation'
import { ConversationList } from './ConversationList'
import styles from './ChatPage.module.css'

export function ChatPage() {
  const { id } = useParams()
  const { role } = useSession()
  const me = useMe()
  const list = useQuery({
    queryKey: ['workspaces'],
    queryFn: getMyWorkspaces,
    refetchInterval: 15000,
  })

  if (list.isPending || me.isPending) return <p className={styles.state}>Загружаем чаты…</p>

  if (list.isError || me.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось загрузить чаты.</p>
        <Button variant="secondary" onClick={() => list.refetch()}>
          Попробовать ещё раз
        </Button>
      </div>
    )
  }

  const pairs = list.data.items
  const basePath = role === 'mentor' ? '/mentor/chat' : '/student/chat'

  if (pairs.length === 0) {
    return (
      <div className={styles.state}>
        <h2 className={styles.stateTitle}>Чатов пока нет</h2>
        <p>Чат появится вместе с парой — после того как заявку примут и подтвердят.</p>
      </div>
    )
  }

  const active = pairs.find((pair) => pair.id === id) ?? pairs[0]

  return (
    <div className={styles.chat}>
      <aside className={clsx(styles.sidebar, id ? styles.hideOnMobile : '')}>
        <ConversationList
          pairs={pairs}
          activeId={active.id}
          myId={me.data.id}
          basePath={basePath}
          partnerRole={role === 'mentor' ? 'Ученик' : 'Наставник'}
        />
      </aside>

      <section className={clsx(styles.conversation, id ? '' : styles.hideOnMobile)}>
        <Conversation
          key={active.id}
          pair={active}
          myId={me.data.id}
          basePath={basePath}
          partnerRole={role === 'mentor' ? 'Ученик' : 'Наставник'}
          workspacePath={role === 'mentor' ? `/mentor/students/${active.id}` : '/student/workspace'}
        />
      </section>
    </div>
  )
}
