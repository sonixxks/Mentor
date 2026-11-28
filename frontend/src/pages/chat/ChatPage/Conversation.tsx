import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { ArrowLeft, Clock, FileText, Send, Video } from 'lucide-react'
import { Fragment, useEffect, useRef, useState } from 'react'
import type { FormEvent, KeyboardEvent } from 'react'
import { Link } from 'react-router-dom'
import { getMessages, markMessagesRead, sendMessage } from '../../../api/messages'
import type { Message } from '../../../api/messages'
import { getWorkspace } from '../../../api/workspaces'
import type { WorkspaceSummary } from '../../../api/workspaces'
import { Avatar } from '../../../components/ui/Avatar'
import { Button } from '../../../components/ui/Button'
import { ekbDate, formatDate, formatTime, todayEkb } from '../../../lib/dates'
import { errorMessage } from '../../../lib/errorMessage'
import styles from './ChatPage.module.css'

const PAGE = 50

type ConversationProps = {
  pair: WorkspaceSummary
  myId: string
  basePath: string
  partnerRole: string
  workspacePath: string
}

function dayTitle(iso: string) {
  const day = ekbDate(iso)
  if (day === todayEkb()) return 'Сегодня'
  if (day === todayEkb(-1)) return 'Вчера'
  return formatDate(iso)
}

function merge(older: Message[], latest: Message[]) {
  const byId = new Map([...older, ...latest].map((message) => [message.id, message]))
  return [...byId.values()].sort((a, b) => a.created_at.localeCompare(b.created_at))
}

export function Conversation({
  pair,
  myId,
  basePath,
  partnerRole,
  workspacePath,
}: ConversationProps) {
  const queryClient = useQueryClient()
  const listRef = useRef<HTMLDivElement>(null)
  const [older, setOlder] = useState<Message[]>([])
  const [hasMore, setHasMore] = useState(true)
  const [loadingOlder, setLoadingOlder] = useState(false)
  const [text, setText] = useState('')

  const name = `${pair.partner.first_name} ${pair.partner.last_name}`.trim()

  const latest = useQuery({
    queryKey: ['messages', pair.id],
    queryFn: () => getMessages(pair.id),
    refetchInterval: 5000,
  })
  const detail = useQuery({
    queryKey: ['workspace', pair.id],
    queryFn: () => getWorkspace(pair.id),
  })

  const messages = merge(older, latest.data ?? [])
  const lastId = messages[messages.length - 1]?.id
  const hasUnread = messages.some(
    (message) => message.type === 'text' && message.sender_id !== myId && !message.is_read,
  )

  useEffect(() => {
    const list = listRef.current
    if (list) list.scrollTop = list.scrollHeight
  }, [lastId])

  useEffect(() => {
    if (!hasUnread) return
    markMessagesRead(pair.id).then(() => {
      queryClient.invalidateQueries({ queryKey: ['messages', pair.id] })
      queryClient.invalidateQueries({ queryKey: ['workspaces'] })
    })
  }, [hasUnread, pair.id, queryClient])

  const send = useMutation({
    mutationFn: () => sendMessage(pair.id, text.trim()),
    onSuccess: () => {
      setText('')
      queryClient.invalidateQueries({ queryKey: ['messages', pair.id] })
    },
  })

  async function loadOlder() {
    const first = messages[0]
    if (!first) return
    setLoadingOlder(true)
    try {
      const page = await getMessages(pair.id, first.created_at)
      setOlder((current) => [...page, ...current])
      if (page.length < PAGE) setHasMore(false)
    } finally {
      setLoadingOlder(false)
    }
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (text.trim() && !send.isPending) send.mutate()
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      if (text.trim() && !send.isPending) send.mutate()
    }
  }

  return (
    <>
      <header className={styles.header}>
        <Link to={basePath} className={styles.back} aria-label="Ко всем чатам">
          <ArrowLeft size={20} />
        </Link>
        <Avatar name={name} size="m" />
        <div className={styles.headerInfo}>
          <span className={styles.itemName}>{name}</span>
          <span className={styles.itemRole}>{partnerRole}</span>
        </div>
        <Button variant="secondary" size="sm" to={workspacePath} className={styles.workspaceLink}>
          Пространство пары
        </Button>
        {detail.data?.jitsi_link ? (
          <a
            href={detail.data.jitsi_link}
            target="_blank"
            rel="noreferrer"
            className={styles.iconButton}
            aria-label="Видеозвонок"
          >
            <Video size={20} />
          </a>
        ) : null}
      </header>

      <div className={styles.messages} ref={listRef}>
        {hasMore && messages.length >= PAGE ? (
          <button
            type="button"
            className={styles.older}
            disabled={loadingOlder}
            onClick={loadOlder}
          >
            {loadingOlder ? 'Загружаем…' : 'Показать раньше'}
          </button>
        ) : null}

        {latest.isPending ? <p className={styles.state}>Загружаем сообщения…</p> : null}

        {latest.isSuccess && messages.length === 0 ? (
          <p className={styles.state}>Сообщений пока нет. Напишите первым.</p>
        ) : null}

        {messages.map((message, index) => {
          const previous = messages[index - 1]
          const newDay = !previous || ekbDate(previous.created_at) !== ekbDate(message.created_at)
          const mine = message.sender_id === myId

          return (
            <Fragment key={message.id}>
              {newDay ? <span className={styles.day}>{dayTitle(message.created_at)}</span> : null}

              {message.type === 'system' ? (
                <span className={styles.system}>
                  <Clock size={14} />
                  {message.text}
                </span>
              ) : (
                <div className={clsx(styles.bubble, mine ? styles.mine : styles.theirs)}>
                  {message.file ? (
                    <a
                      href={message.file.file_url}
                      target="_blank"
                      rel="noreferrer"
                      className={styles.file}
                    >
                      <span className={styles.fileIcon}>
                        <FileText size={18} />
                      </span>
                      <span className={styles.fileInfo}>
                        <span className={styles.fileName}>{message.file.file_name}</span>
                        {message.file.file_size_kb ? (
                          <span className={styles.fileSize}>
                            {message.file.file_size_kb.toLocaleString('ru-RU')} КБ
                          </span>
                        ) : null}
                      </span>
                    </a>
                  ) : null}
                  {message.text ? <p className={styles.text}>{message.text}</p> : null}
                  <span className={styles.meta}>
                    {formatTime(message.created_at)}
                    {mine ? (message.is_read ? ' · прочитано' : ' · отправлено') : ''}
                  </span>
                </div>
              )}
            </Fragment>
          )
        })}
      </div>

      <form className={styles.composer} onSubmit={handleSubmit}>
        <textarea
          className={styles.input}
          rows={1}
          maxLength={4000}
          placeholder="Сообщение"
          aria-label="Сообщение"
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
        />
        <button
          type="submit"
          className={styles.send}
          disabled={!text.trim() || send.isPending}
          aria-label="Отправить"
        >
          <Send size={18} />
        </button>
      </form>

      {send.isError ? (
        <p className={styles.error} role="alert">
          {errorMessage(send.error)}
        </p>
      ) : null}
    </>
  )
}
