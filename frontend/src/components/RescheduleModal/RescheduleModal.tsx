import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { updateMeeting } from '../../api/workspaces'
import type { Meeting } from '../../api/workspaces'
import { ekbDate, ekbTime, fromEkb } from '../../lib/dates'
import { errorMessage } from '../../lib/errorMessage'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Modal } from '../ui/Modal'
import styles from './RescheduleModal.module.css'

type RescheduleModalProps = {
  meeting: Meeting
  partnerName: string
  onClose: () => void
}

export function RescheduleModal({ meeting, partnerName, onClose }: RescheduleModalProps) {
  const queryClient = useQueryClient()
  const [date, setDate] = useState(ekbDate(meeting.start_time))
  const [time, setTime] = useState(ekbTime(meeting.start_time))

  const filled = Boolean(date && time)
  const inPast = filled && Date.parse(fromEkb(date, time)) < Date.now()

  const move = useMutation({
    mutationFn: () => {
      const start = fromEkb(date, time)
      const length = meeting.end_time
        ? Date.parse(meeting.end_time) - Date.parse(meeting.start_time)
        : 0
      return updateMeeting(meeting.id, {
        start_time: start,
        end_time: length ? new Date(Date.parse(start) + length).toISOString() : undefined,
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['calendar'] })
      queryClient.invalidateQueries({ queryKey: ['workspace', meeting.workspace_id] })
      onClose()
    },
  })

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (filled && !inPast) move.mutate()
  }

  return (
    <Modal title="Перенести встречу" onClose={onClose}>
      <form className={styles.form} onSubmit={handleSubmit}>
        <p className={styles.lead}>
          {meeting.title} · {partnerName}
        </p>

        <div className={styles.row}>
          <Input label="Дата" type="date" value={date} onChange={(e) => setDate(e.target.value)} />
          <Input label="Время" type="time" value={time} onChange={(e) => setTime(e.target.value)} />
        </div>

        <p className={styles.hint}>
          Время по Екатеринбургу (UTC+5). {partnerName} получит уведомление о переносе.
        </p>

        {inPast ? <p className={styles.error}>Выберите время в будущем.</p> : null}

        {move.isError ? (
          <p className={styles.error} role="alert">
            {errorMessage(move.error)}
          </p>
        ) : null}

        <div className={styles.actions}>
          <Button variant="secondary" onClick={onClose}>
            Отмена
          </Button>
          <Button type="submit" disabled={!filled || inPast || move.isPending}>
            {move.isPending ? 'Сохраняем…' : 'Перенести'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
