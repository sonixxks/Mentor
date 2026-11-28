import { useState } from 'react'
import type { WorkspaceDetail } from '../../../api/workspaces'
import { MeetingModal } from '../../../components/MeetingModal'
import { RescheduleModal } from '../../../components/RescheduleModal'
import { Button } from '../../../components/ui/Button'
import { formatMeetingTime, minutesBetween } from '../../../lib/dates'
import styles from './MentorStudentPage.module.css'

export function MeetingPlanner({ workspace }: { workspace: WorkspaceDetail }) {
  const [planning, setPlanning] = useState(false)
  const [moving, setMoving] = useState(false)

  const next = workspace.next_meeting
  const minutes = next ? minutesBetween(next.start_time, next.end_time) : null
  const partner = workspace.partner.first_name

  return (
    <section className={styles.card}>
      <h3 className={styles.cardTitle}>Следующая встреча</h3>

      {next ? (
        <div className={styles.nextMeeting}>
          <strong className={styles.meetingTime}>{formatMeetingTime(next.start_time)}</strong>
          <span className={styles.secondary}>
            {next.title}
            {minutes ? ` · ${minutes} мин` : ''}
          </span>
          <div className={styles.meetingRow}>
            <a href={next.jitsi_link} target="_blank" rel="noreferrer" className={styles.join}>
              Подключиться
            </a>
            <Button variant="secondary" size="sm" onClick={() => setMoving(true)}>
              Перенести
            </Button>
          </div>
          <button type="button" className={styles.addLink} onClick={() => setPlanning(true)}>
            + Назначить ещё встречу
          </button>
        </div>
      ) : (
        <>
          <p className={styles.secondary}>Встреча пока не назначена.</p>
          <Button variant="accent" fullWidth onClick={() => setPlanning(true)}>
            Назначить встречу
          </Button>
        </>
      )}

      <p className={styles.hint}>
        Время по Екатеринбургу (UTC+5). Ссылка на звонок постоянная для пары.
      </p>

      {planning ? (
        <MeetingModal
          workspaceId={workspace.id}
          partnerName={partner}
          onClose={() => setPlanning(false)}
        />
      ) : null}

      {moving && next ? (
        <RescheduleModal meeting={next} partnerName={partner} onClose={() => setMoving(false)} />
      ) : null}
    </section>
  )
}
