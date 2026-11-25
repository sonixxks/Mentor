import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { createMeeting } from '../../api/workspaces'
import { fromEkb, todayEkb } from '../../lib/dates'
import { errorMessage } from '../../lib/errorMessage'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Modal } from '../ui/Modal'
import { Select } from '../ui/Select'
import styles from './MeetingModal.module.css'

const DURATIONS = [
  { value: '30', label: '30 минут' },
  { value: '45', label: '45 минут' },
  { value: '60', label: '60 минут' },
  { value: '90', label: '90 минут' },
]

type MeetingModalProps = {
  workspaceId: string
  partnerName: string
  onClose: () => void
}

export function MeetingModal({ workspaceId, partnerName, onClose }: MeetingModalProps) {
  const queryClient = useQueryClient()
  const [title, setTitle] = useState('')
  const [date, setDate] = useState(todayEkb(1))
  const [time, setTime] = useState('18:00')
  const [duration, setDuration] = useState('60')

  const filled = Boolean(date && time)
  const inPast = filled && Date.parse(fromEkb(date, time)) < Date.now()

  const create = useMutation({
    mutationFn: () => {
      const start = fromEkb(date, time)
      const end = new Date(Date.parse(start) + Number(duration) * 60000).toISOString()
      return createMeeting(workspaceId, {
        title: title.trim() || undefined,
        start_time: start,
        end_time: end,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['workspace', workspaceId] })
      queryClient.invalidateQueries({ queryKey: ['calendar'] })
      onClose()
    },
  })

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (filled && !inPast) create.mutate()
  }

  return (
    <Modal title="Назначить встречу" onClose={onClose}>
      <form className={styles.form} onSubmit={handleSubmit}>
        <Input
          label="Тема"
          placeholder="Встреча"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />

        <div className={styles.row}>
          <Input
            label="Дата"
            type="date"
            min={todayEkb()}
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
          <Input label="Время" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
        </div>

        <Select
          label="Длительность"
          options={DURATIONS}
          value={duration}
          onChange={(e) => setDuration(e.target.value)}
        />

        <p className={styles.hint}>
          Время по Екатеринбургу (UTC+5). {partnerName} получит уведомление, звонок пройдёт по
          постоянной ссылке пары.
        </p>

        {inPast ? <p className={styles.error}>Выберите время в будущем.</p> : null}

        {create.isError ? (
          <p className={styles.error} role="alert">
            {errorMessage(create.error)}
          </p>
        ) : null}

        <div className={styles.actions}>
          <Button variant="secondary" onClick={onClose}>
            Отмена
          </Button>
          <Button type="submit" disabled={!filled || inPast || create.isPending}>
            {create.isPending ? 'Сохраняем…' : 'Назначить'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
