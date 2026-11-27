import { useQuery } from '@tanstack/react-query'
import { CheckCircle2, ShieldCheck, Star } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { getMentor, getMentorReviews } from '../../../api/mentors'
import { RequestModal } from '../../../components/RequestModal'
import { Button } from '../../../components/ui/Button'
import { Tag } from '../../../components/ui/Tag'
import {
  EXPERIENCE_LABEL,
  HELP_TOPIC_LABEL,
  STUDENT_LEVEL_LABEL,
  formatPrice,
  formatRating,
  freeSlotsLabel,
  reviewsLabel,
} from '../../../lib/labels'
import { plural } from '../../../lib/plural'
import { summarizeSlots } from '../../../lib/schedule'
import styles from './MentorProfilePage.module.css'

const MATCH_ROWS = [
  ['skills', 'Навыки и интересы'],
  ['goals', 'Цели'],
  ['time', 'Свободное время'],
  ['style', 'Стиль общения'],
] as const

function reviewDate(iso: string) {
  return new Date(iso)
    .toLocaleDateString('ru-RU', { month: 'long', year: 'numeric', timeZone: 'Asia/Yekaterinburg' })
    .replace(' г.', '')
}

export function MentorProfilePage() {
  const { id = '' } = useParams()
  const [requestOpen, setRequestOpen] = useState(false)
  const mentor = useQuery({ queryKey: ['mentor', id], queryFn: () => getMentor(id) })
  const reviews = useQuery({
    queryKey: ['mentor', id, 'reviews'],
    queryFn: () => getMentorReviews(id),
  })

  if (mentor.isPending) return <p className={styles.state}>Загружаем профиль…</p>
  if (mentor.isError) return <p className={styles.state}>Наставник не найден.</p>

  const m = mentor.data
  const name = `${m.first_name} ${m.last_name}`
  const initials = `${m.first_name[0] ?? ''}${m.last_name[0] ?? ''}`
  const noSlots = m.free_slots === 0
  const facts = [
    m.job_title,
    m.company ?? 'работает самостоятельно',
    m.experience ? `опыт ${EXPERIENCE_LABEL[m.experience]}` : null,
  ]
    .filter(Boolean)
    .join(' · ')
  const schedule = summarizeSlots(m.availability_slots)

  return (
    <div className={styles.page}>
      <aside className={styles.left}>
        <section className={styles.card}>
          <div className={styles.portrait}>
            <span className={styles.initials}>{initials}</span>
            <span className={noSlots ? styles.slotsNone : styles.slots}>
              {freeSlotsLabel(m.free_slots)}
            </span>
          </div>

          <h2 className={styles.name}>{name}</h2>
          {m.job_title ? <p className={styles.position}>{m.job_title}</p> : null}

          {m.verification_status === 'verified' ? (
            <span className={styles.verified}>
              <ShieldCheck size={16} />
              Опыт подтверждён куратором
            </span>
          ) : null}

          <span className={styles.rating}>
            <Star size={16} fill="currentColor" className={styles.star} />
            <strong>{formatRating(m.rating_avg)}</strong>
            {m.reviews_count > 0 ? <span>{reviewsLabel(m.reviews_count)}</span> : null}
          </span>

          <Button size="lg" fullWidth disabled={noSlots} onClick={() => setRequestOpen(true)}>
            {noSlots ? 'Мест нет' : 'Отправить заявку'}
          </Button>
        </section>

        <section className={styles.card}>
          <h3 className={styles.cardTitle}>Доступность</h3>
          <dl className={styles.facts}>
            {m.time_limit ? (
              <div>
                <dt>Время</dt>
                <dd>{plural(m.time_limit, ['час', 'часа', 'часов'])} в неделю</dd>
              </div>
            ) : null}
            {schedule ? (
              <div>
                <dt>Когда</dt>
                <dd>{schedule}</dd>
              </div>
            ) : null}
            <div>
              <dt>Цена</dt>
              <dd>{formatPrice(m.hourly_rate, ' за занятие')}</dd>
            </div>
          </dl>
          <p className={styles.hint}>
            Время по Екатеринбургу (UTC+5). Оплата — напрямую наставнику, платформа не участвует в
            расчётах.
          </p>
        </section>
      </aside>

      <div className={styles.right}>
        {m.match ? (
          <section className={styles.match}>
            <div className={styles.matchValue}>
              <strong>{m.match.percent}%</strong>
              <span>совместимость с тобой</span>
            </div>
            <div className={styles.matchRows}>
              {MATCH_ROWS.map(([key, label]) => (
                <div key={key} className={styles.matchRow}>
                  <div className={styles.matchHead}>
                    <span>{label}</span>
                    <strong>{m.match!.breakdown[key]}%</strong>
                  </div>
                  <span className={styles.bar}>
                    <span style={{ width: `${m.match!.breakdown[key]}%` }} />
                  </span>
                </div>
              ))}
            </div>
          </section>
        ) : null}

        <section className={styles.card}>
          <h3 className={styles.cardTitle}>О себе</h3>
          {m.bio ? <p className={styles.text}>{m.bio}</p> : null}
          <p className={styles.facts2}>{facts}</p>

          <span className={styles.label}>Навыки</span>
          <div className={styles.tags}>
            {m.current_skills.map((item) => (
              <Tag key={item.skill}>
                {item.skill}
                {item.level === 3 ? ' · эксперт' : ''}
              </Tag>
            ))}
          </div>

          {m.help_topics.length > 0 ? (
            <>
              <span className={styles.label}>С чем помогает</span>
              <div className={styles.tags}>
                {m.help_topics.map((topic) => (
                  <Tag key={topic}>{HELP_TOPIC_LABEL[topic] ?? topic}</Tag>
                ))}
              </div>
            </>
          ) : null}

          {m.target_student_levels.length > 0 ? (
            <>
              <span className={styles.label}>С кем работает</span>
              <div className={styles.tags}>
                {m.target_student_levels.map((level) => (
                  <Tag key={level}>{STUDENT_LEVEL_LABEL[level] ?? level}</Tag>
                ))}
              </div>
            </>
          ) : null}
        </section>

        <div className={styles.lower}>
          <section className={styles.card}>
            <div className={styles.cardHead}>
              <h3 className={styles.cardTitle}>Авторские программы</h3>
              <span className={styles.muted}>или открытый запрос — план составите вместе</span>
            </div>
            {m.service_packages.length === 0 ? (
              <p className={styles.muted}>
                Программ пока нет. Отправь открытый запрос — цели и план обсудите на первой встрече.
              </p>
            ) : (
              m.service_packages.map((program) => (
                <article key={program.id} className={styles.program}>
                  <div className={styles.programHead}>
                    <h4>{program.title}</h4>
                    <span className={styles.programMeta}>
                      {program.session_count
                        ? `${plural(program.session_count, ['занятие', 'занятия', 'занятий'])} · `
                        : ''}
                      {formatPrice(program.price, ' за занятие')}
                    </span>
                  </div>
                  {program.description ? (
                    <p className={styles.text}>{program.description}</p>
                  ) : null}
                  {program.result ? (
                    <p className={styles.result}>
                      <CheckCircle2 size={16} />
                      <span>
                        <b>Результат:</b> {program.result}
                      </span>
                    </p>
                  ) : null}
                </article>
              ))
            )}
          </section>

          <section className={styles.card}>
            <div className={styles.cardHead}>
              <h3 className={styles.cardTitle}>Отзывы учеников</h3>
              {reviews.data && reviews.data.total > 0 ? (
                <span className={styles.muted}>Всего {reviews.data.total}</span>
              ) : null}
            </div>
            {reviews.isPending ? <p className={styles.muted}>Загружаем…</p> : null}
            {reviews.data?.items.length === 0 ? (
              <p className={styles.muted}>Отзывов пока нет.</p>
            ) : null}
            {reviews.data?.items.map((review) => (
              <article key={review.id} className={styles.review}>
                <span className={styles.stars}>{'★'.repeat(review.rating)}</span>
                {review.comment ? <p className={styles.text}>«{review.comment}»</p> : null}
                <span className={styles.muted}>
                  {review.author_first_name} · {reviewDate(review.created_at)}
                </span>
              </article>
            ))}
          </section>
        </div>
      </div>

      {requestOpen ? <RequestModal mentor={m} onClose={() => setRequestOpen(false)} /> : null}
    </div>
  )
}
