import { useMemo, useState } from 'react'
import { Check, Clock, Search, Star, X } from 'lucide-react'
import { Avatar } from '../../components/ui/Avatar/Avatar'
import styles from './StudentCatalog.module.css'
import { Input } from '../../components/ui/Input/Input'
import { Select } from '../../components/ui/Select/Select'

type Mentor = {
  id: number
  name: string
  position: string
  experience: string
  rating: number
  reviews: number
  match: number
  skills: string[]
  hours: number
  schedule: string
  price: number
  freePlaces: boolean
}

const mentors: Mentor[] = [
  {
    id: 1,
    name: 'Данил Королёв',
    position: 'ML-инженер',
    experience: '5 лет',
    rating: 4.9,
    reviews: 38,
    match: 94,
    skills: ['Python', 'ML'],
    hours: 1.5,
    schedule: 'Вт, Чт 19–21',
    price: 1000,
    freePlaces: true,
  },
  {
    id: 2,
    name: 'Мария Петрова',
    position: 'Data Analyst',
    experience: '6 лет',
    rating: 4.8,
    reviews: 24,
    match: 91,
    skills: ['SQL', 'Статистика'],
    hours: 2,
    schedule: 'Сб 10–14',
    price: 1200,
    freePlaces: true,
  },
  {
    id: 3,
    name: 'Виктор Соколов',
    position: 'CV-инженер',
    experience: '8 лет',
    rating: 4.7,
    reviews: 17,
    match: 87,
    skills: ['Computer Vision'],
    hours: 1,
    schedule: 'Ср 20–21',
    price: 1500,
    freePlaces: false,
  },
  {
    id: 4,
    name: 'Ольга Лебедева',
    position: 'Тимлид DS',
    experience: '10 лет',
    rating: 5.0,
    reviews: 41,
    match: 84,
    skills: ['MLOps', 'Карьера'],
    hours: 2,
    schedule: 'Пн, Пт 8–9',
    price: 0,
    freePlaces: true,
  },
  {
    id: 5,
    name: 'Игорь Новиков',
    position: 'Backend + ML',
    experience: '4 года',
    rating: 4.6,
    reviews: 9,
    match: 79,
    skills: ['Python', 'API'],
    hours: 3,
    schedule: 'Вт–Чт 18–19',
    price: 900,
    freePlaces: true,
  },
  {
    id: 6,
    name: 'Анна Егорова',
    position: 'Product Analyst',
    experience: '5 лет',
    rating: 4.8,
    reviews: 22,
    match: 76,
    skills: ['A/B-тесты'],
    hours: 1.5,
    schedule: 'Пн 19–20:30',
    price: 0,
    freePlaces: true,
  },
]

const formatPrice = (price: number) =>
  price === 0 ? 'Бесплатно' : `${price.toLocaleString('ru-RU')} ₽/занятие`

type ChipProps = {
  active: boolean
  children: React.ReactNode
  onClick: () => void
}

function Chip({ active, children, onClick }: ChipProps) {
  return (
    <button
      type="button"
      className={`${styles.chip} ${active ? styles.chipActive : ''}`}
      onClick={onClick}
    >
      {children}
    </button>
  )
}

export function StudentCatalog() {
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState('match')
  const [selectedSkills, setSelectedSkills] = useState<string[]>([])
  const [priceFilter, setPriceFilter] = useState('any')
  const [hoursFilter, setHoursFilter] = useState('any')
  const [minMatch, setMinMatch] = useState(70)
  const [ratingFilter, setRatingFilter] = useState('any')
  const [onlyFreePlaces, setOnlyFreePlaces] = useState(true)
  const [compared, setCompared] = useState<number[]>([])
  const [sentRequests, setSentRequests] = useState<number[]>([])
  const [showComparison, setShowComparison] = useState(false)

  const filteredMentors = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase()

    const result = mentors.filter((mentor) => {
      const matchesQuery =
        !normalizedQuery ||
        mentor.name.toLowerCase().includes(normalizedQuery) ||
        mentor.position.toLowerCase().includes(normalizedQuery) ||
        mentor.skills.some((skill) => skill.toLowerCase().includes(normalizedQuery))

      const matchesSkills =
        selectedSkills.length === 0 ||
        selectedSkills.every((skill) => mentor.skills.includes(skill))

      const matchesPrice =
        priceFilter === 'any' ||
        (priceFilter === 'free' && mentor.price === 0) ||
        (priceFilter === '1500' && mentor.price <= 1500) ||
        (priceFilter === '3000' && mentor.price <= 3000)

      const matchesHours =
        hoursFilter === 'any' ||
        (hoursFilter === '1-2' && mentor.hours >= 1 && mentor.hours <= 2) ||
        (hoursFilter === '3-5' && mentor.hours >= 3 && mentor.hours <= 5) ||
        (hoursFilter === '5+' && mentor.hours > 5)

      const matchesRating =
        ratingFilter === 'any' ||
        (ratingFilter === '4.5' && mentor.rating >= 4.5) ||
        (ratingFilter === '4.8' && mentor.rating >= 4.8)

      return (
        matchesQuery &&
        matchesSkills &&
        matchesPrice &&
        matchesHours &&
        mentor.match >= minMatch &&
        matchesRating &&
        (!onlyFreePlaces || mentor.freePlaces)
      )
    })

    if (sort === 'match') result.sort((a, b) => b.match - a.match)
    if (sort === 'rating') result.sort((a, b) => b.rating - a.rating)
    if (sort === 'price') result.sort((a, b) => a.price - b.price)

    return result
  }, [
    query,
    sort,
    selectedSkills,
    priceFilter,
    hoursFilter,
    minMatch,
    ratingFilter,
    onlyFreePlaces,
  ])

  const toggleSkill = (skill: string) => {
    setSelectedSkills((current) =>
      current.includes(skill) ? current.filter((item) => item !== skill) : [...current, skill],
    )
  }

  const toggleCompare = (id: number) => {
    setCompared((current) => {
      if (current.includes(id)) {
        setShowComparison(false)
        return current.filter((item) => item !== id)
      }

      if (current.length >= 3) return current
      return [...current, id]
    })
  }

  const sendRequest = (id: number) => {
    setSentRequests((current) => (current.includes(id) ? current : [...current, id]))
  }

  const resetFilters = () => {
    setQuery('')
    setSort('match')
    setSelectedSkills([])
    setPriceFilter('any')
    setHoursFilter('any')
    setMinMatch(70)
    setRatingFilter('any')
    setOnlyFreePlaces(false)
  }

  const comparedMentors = mentors.filter((mentor) => compared.includes(mentor.id))

  return (
    <div className={styles.catalog}>
      <div className={styles.layout}>
        <section className={styles.results}>
          <div className={styles.toolbar}>
            <div className={styles.searchWrapper}>
              <Search size={19} className={styles.searchIcon} />
              <Input
                className={styles.searchInput}
                placeholder="Навык, имя или компания"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
              />
            </div>

            <Select
              className={styles.sortSelect}
              value={sort}
              onChange={(event) => setSort(event.target.value)}
              options={[
                { value: 'match', label: 'По совпадению' },
                { value: 'rating', label: 'По рейтингу' },
                { value: 'price', label: 'По цене' },
              ]}
            />
          </div>

          <p className={styles.resultCount}>Найдено {filteredMentors.length} наставников</p>

          {filteredMentors.length > 0 ? (
            <div className={styles.grid}>
              {filteredMentors.map((mentor) => {
                const isCompared = compared.includes(mentor.id)
                const isSent = sentRequests.includes(mentor.id)

                return (
                  <article className={styles.mentorCard} key={mentor.id}>
                    <div className={styles.cardHeader}>
                      <Avatar name={mentor.name} size="l" />

                      <div className={styles.mentorInfo}>
                        <h3>{mentor.name}</h3>
                        <p>
                          {mentor.position} · {mentor.experience}
                        </p>
                        <span className={styles.rating}>
                          <Star size={14} fill="currentColor" />
                          <strong>{mentor.rating.toFixed(1).replace('.', ',')}</strong>
                          <span>· {mentor.reviews} отзывов</span>
                        </span>
                      </div>
                    </div>

                    <div className={styles.tags}>
                      <span className={styles.matchTag}>{mentor.match}% совпадение</span>
                      {mentor.skills.map((skill) => (
                        <span className={styles.skillTag} key={skill}>
                          {skill}
                        </span>
                      ))}
                    </div>

                    <div className={styles.cardMeta}>
                      <span>
                        <Clock size={15} />
                        {String(mentor.hours).replace('.', ',')} ч/нед · {mentor.schedule}
                      </span>
                      <strong>{formatPrice(mentor.price)}</strong>
                    </div>

                    <div className={styles.cardActions}>
                      <button
                        type="button"
                        className={styles.requestButton}
                        onClick={() => sendRequest(mentor.id)}
                        disabled={isSent}
                      >
                        {isSent ? (
                          <>
                            <Check size={15} /> Заявка отправлена
                          </>
                        ) : (
                          'Отправить заявку'
                        )}
                      </button>

                      <button
                        type="button"
                        className={`${styles.compareButton} ${
                          isCompared ? styles.compareButtonActive : ''
                        }`}
                        onClick={() => toggleCompare(mentor.id)}
                        disabled={!isCompared && compared.length >= 3}
                      >
                        {isCompared ? (
                          <>
                            <Check size={15} /> В сравнении
                          </>
                        ) : (
                          'Сравнить'
                        )}
                      </button>
                    </div>
                  </article>
                )
              })}
            </div>
          ) : (
            <div className={styles.empty}>
              <Search size={30} />
              <h3>Наставники не найдены</h3>
              <p>Попробуй изменить параметры поиска.</p>
              <button type="button" onClick={resetFilters}>
                Сбросить фильтры
              </button>
            </div>
          )}
        </section>

        <aside className={styles.filters}>
          <div className={styles.filterHeading}>
            <h2>Фильтры</h2>
            <button type="button" onClick={resetFilters}>
              Сбросить
            </button>
          </div>

          <div className={styles.filterGroup}>
            <h3>Навыки</h3>
            <div className={styles.chips}>
              {['Data Science', 'Python', 'ML', 'SQL'].map((skill) => (
                <Chip
                  key={skill}
                  active={selectedSkills.includes(skill)}
                  onClick={() => toggleSkill(skill)}
                >
                  {skill}
                  {selectedSkills.includes(skill) && ' ×'}
                </Chip>
              ))}
              <Chip active={false} onClick={() => setSelectedSkills([])}>
                + ещё
              </Chip>
            </div>
          </div>

          <div className={styles.filterGroup}>
            <h3>Стоимость занятия</h3>
            <div className={styles.chips}>
              <Chip active={priceFilter === 'any'} onClick={() => setPriceFilter('any')}>
                Любая
              </Chip>
              <Chip active={priceFilter === 'free'} onClick={() => setPriceFilter('free')}>
                Бесплатно
              </Chip>
              <Chip active={priceFilter === '1500'} onClick={() => setPriceFilter('1500')}>
                до 1 500 ₽
              </Chip>
              <Chip active={priceFilter === '3000'} onClick={() => setPriceFilter('3000')}>
                до 3 000 ₽
              </Chip>
            </div>
            <p className={styles.filterHint}>Оплата — напрямую наставнику, вне платформы</p>
          </div>

          <div className={styles.filterGroup}>
            <h3>Время в неделю</h3>
            <div className={styles.chips}>
              <Chip active={hoursFilter === '1-2'} onClick={() => setHoursFilter('1-2')}>
                1–2 ч
              </Chip>
              <Chip active={hoursFilter === '3-5'} onClick={() => setHoursFilter('3-5')}>
                3–5 ч
              </Chip>
              <Chip active={hoursFilter === '5+'} onClick={() => setHoursFilter('5+')}>
                5+ ч
              </Chip>
            </div>
            <button
              type="button"
              className={styles.resetSmall}
              onClick={() => setHoursFilter('any')}
            >
              Не важно
            </button>
          </div>

          <div className={styles.filterGroup}>
            <div className={styles.rangeHeading}>
              <h3>Минимальное совпадение</h3>
              <strong>от {minMatch}%</strong>
            </div>
            <input
              className={styles.range}
              type="range"
              min="0"
              max="100"
              step="5"
              value={minMatch}
              onChange={(event) => setMinMatch(Number(event.target.value))}
            />
          </div>

          <div className={styles.filterGroup}>
            <h3>Рейтинг</h3>
            <div className={styles.chips}>
              <Chip active={ratingFilter === 'any'} onClick={() => setRatingFilter('any')}>
                Любой
              </Chip>
              <Chip active={ratingFilter === '4.5'} onClick={() => setRatingFilter('4.5')}>
                от 4,5
              </Chip>
              <Chip active={ratingFilter === '4.8'} onClick={() => setRatingFilter('4.8')}>
                от 4,8
              </Chip>
            </div>
          </div>

          <label className={styles.switchRow}>
            <span>Есть свободные места</span>
            <input
              type="checkbox"
              checked={onlyFreePlaces}
              onChange={(event) => setOnlyFreePlaces(event.target.checked)}
            />
            <span className={styles.switch} />
          </label>
        </aside>
      </div>

      {compared.length > 0 && (
        <div className={styles.compareBar}>
          <div className={styles.compareAvatars}>
            {comparedMentors.map((mentor) => (
              <span key={mentor.id} title={mentor.name}>
                {mentor.name
                  .split(' ')
                  .map((part) => part[0])
                  .join('')
                  .slice(0, 2)}
              </span>
            ))}
          </div>

          <strong>Выбрано {compared.length} из 3</strong>

          <button
            type="button"
            className={styles.compareMainButton}
            disabled={compared.length < 2}
            onClick={() => setShowComparison((current) => !current)}
          >
            {showComparison ? 'Скрыть сравнение' : 'Сравнить'}
          </button>

          <button
            type="button"
            className={styles.clearCompare}
            onClick={() => {
              setCompared([])
              setShowComparison(false)
            }}
            aria-label="Очистить сравнение"
          >
            <X size={17} />
          </button>
        </div>
      )}

      {showComparison && comparedMentors.length >= 2 && (
        <section className={styles.comparison}>
          <div className={styles.comparisonHeading}>
            <h3>Сравнение наставников</h3>
            <button type="button" onClick={() => setShowComparison(false)}>
              <X size={18} />
            </button>
          </div>

          <div className={styles.comparisonGrid}>
            {comparedMentors.map((mentor) => (
              <div key={mentor.id}>
                <strong>{mentor.name}</strong>
                <span>{mentor.position}</span>
                <span>Совпадение: {mentor.match}%</span>
                <span>Рейтинг: {mentor.rating.toFixed(1).replace('.', ',')}</span>
                <span>{formatPrice(mentor.price)}</span>
                <span>{mentor.hours} ч/нед</span>
              </div>
            ))}
          </div>
        </section>
      )}

      <p className={styles.demoNote}>
        Демонстрационные данные. Заявки и сравнение пока работают локально.
      </p>
    </div>
  )
}
