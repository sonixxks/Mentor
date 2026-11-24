import { Clock, Star } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { MentorSummary } from '../../api/mentors'
import { EXPERIENCE_LABEL, formatPrice, formatRating, reviewsLabel } from '../../lib/labels'
import { Avatar } from '../ui/Avatar'
import { Badge } from '../ui/Badge'
import { Button } from '../ui/Button'
import { Card } from '../ui/Card'
import { Tag } from '../ui/Tag'
import styles from './MentorCard.module.css'

type MentorCardProps = {
  mentor: MentorSummary
  onRequest: () => void
}

export function MentorCard({ mentor, onRequest }: MentorCardProps) {
  const name = `${mentor.first_name} ${mentor.last_name}`
  const profileUrl = `/student/mentors/${mentor.id}`
  const experience = mentor.experience ? EXPERIENCE_LABEL[mentor.experience] : null
  const subtitle = [mentor.job_title, experience].filter(Boolean).join(' · ')
  const noSlots = mentor.free_slots === 0

  return (
    <Card className={styles.card}>
      <div className={styles.head}>
        <Link to={profileUrl} className={styles.avatarLink} aria-label={`Профиль: ${name}`}>
          <Avatar name={name} size="m" />
        </Link>
        <div className={styles.info}>
          <Link to={profileUrl} className={styles.name}>
            {name}
          </Link>
          {subtitle ? <span className={styles.subtitle}>{subtitle}</span> : null}
          <span className={styles.rating}>
            <Star size={14} fill="currentColor" className={styles.star} />
            {formatRating(mentor.rating_avg)}
            {mentor.reviews_count > 0 ? ` · ${reviewsLabel(mentor.reviews_count)}` : ''}
          </span>
        </div>
      </div>

      <div className={styles.tags}>
        {mentor.match ? <Badge variant="match">{mentor.match.percent}% совпадение</Badge> : null}
        {mentor.current_skills.slice(0, 3).map((item) => (
          <Tag key={item.skill}>{item.skill}</Tag>
        ))}
      </div>

      <div className={styles.meta}>
        <span className={styles.hours}>
          <Clock size={15} />
          {mentor.time_limit ? `${mentor.time_limit} ч/нед` : 'Время не указано'}
        </span>
        <strong>{formatPrice(mentor.hourly_rate)}</strong>
      </div>

      <Button size="sm" fullWidth onClick={onRequest} disabled={noSlots}>
        {noSlots ? 'Мест нет' : 'Отправить заявку'}
      </Button>
    </Card>
  )
}
