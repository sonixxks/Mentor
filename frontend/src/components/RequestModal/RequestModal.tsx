import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CheckCircle2 } from 'lucide-react'
import { useState } from 'react'
import type { FormEvent } from 'react'
import { getMentor } from '../../api/mentors'
import type { ServicePackage } from '../../api/mentors'
import { createRequest } from '../../api/requests'
import { errorMessage } from '../../lib/errorMessage'
import { formatPrice } from '../../lib/labels'
import { plural } from '../../lib/plural'
import { Button } from '../ui/Button'
import { ChoiceCard } from '../ui/ChoiceCard'
import { Modal } from '../ui/Modal'
import { Textarea } from '../ui/Textarea'
import styles from './RequestModal.module.css'

export type RequestTarget = {
  id: string
  first_name: string
  last_name: string
}

type RequestModalProps = {
  mentor: RequestTarget
  onClose: () => void
}

function programMeta(program: ServicePackage) {
  return [
    program.session_count ? plural(program.session_count, ['занятие', 'занятия', 'занятий']) : null,
    formatPrice(program.price, ' за занятие'),
  ]
    .filter(Boolean)
    .join(' · ')
}

export function RequestModal({ mentor, onClose }: RequestModalProps) {
  const queryClient = useQueryClient()
  const [programId, setProgramId] = useState<string | null>(null)
  const [letter, setLetter] = useState('')

  const details = useQuery({
    queryKey: ['mentor', mentor.id],
    queryFn: () => getMentor(mentor.id),
  })
  const programs = details.data?.service_packages ?? []

  const send = useMutation({
    mutationFn: () =>
      createRequest({
        mentor_id: mentor.id,
        service_package_id: programId ?? undefined,
        cover_letter: letter.trim() || undefined,
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['requests'] }),
  })

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    send.mutate()
  }

  if (send.isSuccess) {
    return (
      <Modal title="Заявка отправлена" onClose={onClose}>
        <div className={styles.success}>
          <CheckCircle2 size={48} className={styles.successIcon} />
          <p>
            {mentor.first_name} получит уведомление и ответит в ближайшие дни. Следить за статусом
            можно в «Моих заявках».
          </p>
          <div className={styles.actions}>
            <Button variant="secondary" onClick={onClose}>
              Закрыть
            </Button>
            <Button to="/student/requests">Мои заявки</Button>
          </div>
        </div>
      </Modal>
    )
  }

  return (
    <Modal title="Заявка наставнику" onClose={onClose}>
      <form className={styles.form} onSubmit={handleSubmit}>
        <p className={styles.lead}>
          {mentor.first_name} {mentor.last_name}
        </p>

        <fieldset className={styles.options}>
          <legend className={styles.label}>С чем идёшь</legend>
          <ChoiceCard
            title="Открытый запрос"
            description="Цели и план обсудите на первой встрече"
            selected={programId === null}
            onClick={() => setProgramId(null)}
          />
          {programs.map((program) => (
            <ChoiceCard
              key={program.id}
              title={program.title}
              description={programMeta(program)}
              selected={programId === program.id}
              onClick={() => setProgramId(program.id)}
            />
          ))}
          {details.isPending ? <p className={styles.muted}>Загружаем программы…</p> : null}
        </fieldset>

        <Textarea
          label="Письмо наставнику"
          rows={5}
          maxLength={1000}
          placeholder="Чего хочешь достичь и почему выбираешь этого наставника. Необязательно, но так ответ придёт быстрее"
          value={letter}
          onChange={(e) => setLetter(e.target.value)}
        />

        <p className={styles.muted}>
          Одновременно можно держать до 3 заявок. Когда примешь одну, остальные отменятся сами.
        </p>

        {send.isError ? (
          <p className={styles.error} role="alert">
            {errorMessage(send.error, 'Не удалось отправить заявку. Попробуй ещё раз.')}
          </p>
        ) : null}

        <div className={styles.actions}>
          <Button variant="secondary" onClick={onClose}>
            Отмена
          </Button>
          <Button type="submit" disabled={send.isPending}>
            {send.isPending ? 'Отправляем…' : 'Отправить заявку'}
          </Button>
        </div>
      </form>
    </Modal>
  )
}
