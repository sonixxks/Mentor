import { useQueries } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { Search } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { getMessages } from '../../../api/messages'
import type { Message } from '../../../api/messages'
import type { WorkspaceSummary } from '../../../api/workspaces'
import { Avatar } from '../../../components/ui/Avatar'
import { Input } from '../../../components/ui/Input'
import { ekbDate, formatDate, formatTime, todayEkb } from '../../../lib/dates'
import styles from './ChatPage.module.css'

type ConversationListProps = {
  pairs: WorkspaceSummary[]
  activeId: string
  myId: string
  basePath: string
  partnerRole: string
}

function preview(message: Message | undefined, myId: string) {
  if (!message) return 'Сообщений пока нет'
  const body = message.text ?? (message.file ? `Файл: ${message.file.file_name}` : '')
  return message.sender_id === myId ? `Вы: ${body}` : body
}

function shortTime(iso: string) {
  return ekbDate(iso) === todayEkb() ? formatTime(iso) : formatDate(iso)
}

export function ConversationList({
  pairs,
  activeId,
  myId,
  basePath,
  partnerRole,
}: ConversationListProps) {
  const [query, setQuery] = useState('')

  const last = useQueries({
    queries: pairs.map((pair) => ({
      queryKey: ['messages', pair.id, 'last'],
      queryFn: () => getMessages(pair.id, undefined, 1),
      refetchInterval: 15000,
    })),
  })

  const search = query.trim().toLowerCase()

  return (
    <>
      <div className={styles.search}>
        <Search size={18} className={styles.searchIcon} />
        <Input
          placeholder="Поиск"
          aria-label="Поиск по чатам"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
      </div>

      <div className={styles.list}>
        {pairs.map((pair, index) => {
          const name = `${pair.partner.first_name} ${pair.partner.last_name}`.trim()
          if (search && !name.toLowerCase().includes(search)) return null
          const message = last[index]?.data?.[0]

          return (
            <Link
              key={pair.id}
              to={`${basePath}/${pair.id}`}
              className={clsx(styles.item, pair.id === activeId ? styles.itemActive : '')}
              aria-current={pair.id === activeId ? 'page' : undefined}
            >
              <Avatar name={name} size="m" />
              <span className={styles.itemInfo}>
                <span className={styles.itemTop}>
                  <span className={styles.itemName}>{name}</span>
                  {message ? (
                    <span className={styles.itemTime}>{shortTime(message.created_at)}</span>
                  ) : null}
                </span>
                <span className={styles.itemRole}>{partnerRole}</span>
                <span className={styles.itemBottom}>
                  <span className={styles.itemPreview}>{preview(message, myId)}</span>
                  {pair.unread_messages ? (
                    <span className={styles.unread}>{pair.unread_messages}</span>
                  ) : null}
                </span>
              </span>
            </Link>
          )
        })}
      </div>
    </>
  )
}
