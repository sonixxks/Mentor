import { useMutation, useQuery } from '@tanstack/react-query'
import { ChevronRight, CircleHelp, Search, SlidersHorizontal } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router-dom'
import { getRecommendations } from '../../api/mentors'
import { updateMyProfile } from '../../api/profiles'
import type { ProfileResponse } from '../../api/profiles'
import { MentorCard } from '../../components/MentorCard'
import { Avatar } from '../../components/ui/Avatar'
import { Badge } from '../../components/ui/Badge'
import { Button } from '../../components/ui/Button'
import { useMyProfile } from '../../hooks/useMyProfile'
import { EXPERIENCE_LABEL, formatRating } from '../../lib/labels'
import { plural } from '../../lib/plural'
import { GOALS } from '../onboarding/goals'
import styles from './StudentCabinet.module.css'

const MATCH_ROWS = [
  ['skills', 'Навыки'],
  ['goals', 'Цели'],
  ['time', 'Время'],
  ['style', 'Стиль'],
] as const

/** Сколько карточек видно сразу, не считая лучшего совпадения */
const FIRST_VISIBLE = 4

/** Анкета → «Python, SQL · 2 ч в неделю · найти стажировку или работу за 3–6 мес.» */
function profileSummary(profile?: ProfileResponse) {
  if (!profile) return ''
  const goal = GOALS.find((item) => item.value === profile.main_goals?.[0])?.title
  const min = profile.min_duration_months
  const max = profile.max_duration_months
  const months = min && max ? (min === max ? `за ${min} мес.` : `за ${min}–${max} мес.`) : ''

  return [
    profile.target_skills?.join(', '),
    profile.time_limit ? `${profile.time_limit} ч в неделю` : null,
    goal ? `${goal.toLowerCase()} ${months}`.trim() : null,
  ]
    .filter(Boolean)
    .join(' · ')
}

export function StudentCabinet() {
  const [showAll, setShowAll] = useState(false)
  const profile = useMyProfile()
  const recommendations = useQuery({
    queryKey: ['recommendations'],
    queryFn: getRecommendations,
  })
  const curatorHelp = useMutation({
    mutationFn: () => updateMyProfile({ needs_curator_help: true }),
  })

  if (recommendations.isPending) {
    return <p className={styles.state}>Подбираем наставников…</p>
  }

  if (recommendations.isError) {
    return (
      <div className={styles.state}>
        <p>Не удалось загрузить подборку.</p>
        <Button variant="secondary" onClick={() => recommendations.refetch()}>
          Попробовать ещё раз
        </Button>
      </div>
    )
  }

  const items = recommendations.data
  const helpRequested = curatorHelp.isSuccess || profile.data?.needs_curator_help === true

  if (items.length === 0) {
    const skills = profile.data?.target_skills?.join(', ')

    return (
      <div className={styles.content}>
        <section className={styles.emptyState}>
          <div className={styles.emptyIllustration}>
            <div className={styles.searchOrbit}>
              <span className={styles.orbitDot} />
              <span className={styles.orbitDotSecond} />
              <div className={styles.searchIcon}>
                <Search size={34} />
              </div>
            </div>
          </div>

          <h2>Пока нет подходящих наставников</h2>

          <p className={styles.emptyDescription}>
            {skills ? `По ${skills} и твоему расписанию` : 'По твоей анкете'} совпадений выше 60%
            пока нет. Новые наставники регистрируются каждую неделю — сообщим, как только появится
            подходящий.
          </p>

          <div className={styles.emptyActions}>
            <button
              type="button"
              className={styles.outlineButton}
              onClick={() => curatorHelp.mutate()}
              disabled={helpRequested || curatorHelp.isPending}
            >
              <CircleHelp size={18} />
              {helpRequested ? 'Куратор получил запрос' : 'Запросить помощь куратора'}
            </button>

            <Link to="/student/catalog" className={styles.outlineButton}>
              Смягчить фильтры в каталоге
            </Link>
          </div>
          {/* TODO: неделя 5 — «Уточнить анкету», когда появится страница профиля */}
        </section>
      </div>
    )
  }

  const [best, ...rest] = items
  const visible = showAll ? rest : rest.slice(0, FIRST_VISIBLE)
  const bestName = `${best.first_name} ${best.last_name}`
  const bestUrl = `/student/mentors/${best.id}`
  const bestFacts = [
    best.job_title,
    best.experience ? `опыт ${EXPERIENCE_LABEL[best.experience]}` : null,
    `рейтинг ${formatRating(best.rating_avg)}`,
    best.time_limit ? `${best.time_limit} ч в неделю` : null,
  ]
    .filter(Boolean)
    .join(' · ')

  return (
    <div className={styles.content}>
      <section className={styles.heading}>
        <div>
          <h2>{plural(items.length, ['наставник', 'наставника', 'наставников'])} под твои цели</h2>
          <p>{profileSummary(profile.data)}</p>
        </div>

        <div className={styles.headingActions}>
          {/* TODO: неделя 5 — «Уточнить анкету», когда появится страница профиля */}
          <Button variant="secondary" size="md" to="/student/catalog">
            <SlidersHorizontal size={16} />
            Каталог с фильтрами
          </Button>
        </div>
      </section>

      <section className={styles.featured}>
        <Link to={bestUrl} className={styles.featuredAvatar} aria-label={`Профиль: ${bestName}`}>
          <Avatar name={bestName} size="l" />
        </Link>

        <div className={styles.featuredInfo}>
          <Badge variant="match">Лучшее совпадение</Badge>
          <Link to={bestUrl} className={styles.featuredName}>
            {bestName}
          </Link>
          <p>{bestFacts}</p>

          <div className={styles.featuredActions}>
            {/* TODO: шаг 5 — открыть окно заявки */}
            <Button>Отправить заявку</Button>
            <Button variant="secondary" to={bestUrl}>
              Профиль
            </Button>
          </div>
        </div>

        {best.match ? (
          <div className={styles.matchPanel}>
            <div className={styles.matchValue}>
              <strong>{best.match.percent}%</strong>
              <span>совместимость</span>
            </div>

            {MATCH_ROWS.map(([key, label]) => (
              <div className={styles.matchRow} key={key}>
                <span>{label}</span>
                <div className={styles.progress}>
                  <span style={{ width: `${best.match!.breakdown[key]}%` }} />
                </div>
              </div>
            ))}
          </div>
        ) : null}
      </section>

      {visible.length > 0 ? (
        <section className={styles.mentorGrid}>
          {visible.map((mentor) => (
            <MentorCard
              key={mentor.id}
              mentor={mentor}
              onRequest={() => {
                
              }}
            />
          ))}
        </section>
      ) : null}

      {!showAll && rest.length > FIRST_VISIBLE ? (
        <button type="button" className={styles.showMore} onClick={() => setShowAll(true)}>
          Показать ещё {rest.length - FIRST_VISIBLE} из подборки <ChevronRight size={16} />
        </button>
      ) : null}
    </div>
  )
}
