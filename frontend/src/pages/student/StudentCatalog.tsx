import { useMemo, useState } from 'react'
import { Check, Clock, Search, Star, X } from 'lucide-react'
import { Avatar } from '../../components/ui/Avatar'
import styles from './StudentCatalog.module.css'
import { Input } from '../../components/ui/Input'
import { Select } from '../../components/ui/Select'
import { Slider } from '../../components/ui/Slider'
import { StudentMentorProfile } from './StudentMentorProfile'

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
  const [priceMin, setPriceMin] = useState(0)
  const [priceMax, setPriceMax] = useState(10000)
  const [hoursLimit, setHoursLimit] = useState(1)
  const [minMatch, setMinMatch] = useState(70)
  const [ratingMin, setRatingMin] = useState(1)
  const [onlyFreePlaces, setOnlyFreePlaces] = useState(true)
  const [compared, setCompared] = useState<number[]>([])
  const [sentRequests, setSentRequests] = useState<number[]>([])
  const [showComparison, setShowComparison] = useState(false)
  const [selectedMentor, setSelectedMentor] = useState<Mentor | null>(null)

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

      const matchesPrice = mentor.price >= priceMin && mentor.price <= priceMax

      // time_limit наставника — максимум часов, которые он готов уделять ученику.
      // Значение 10 в анкете означает «10 и более».
      const matchesHours = mentor.hours >= hoursLimit

      const matchesRating = mentor.rating >= ratingMin

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
    priceMin,
    priceMax,
    hoursLimit,
    minMatch,
    ratingMin,
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
    setPriceMin(0)
    setPriceMax(10000)
    setHoursLimit(1)
    setMinMatch(70)
    setRatingMin(1)
    setOnlyFreePlaces(false)
  }

  const comparedMentors = mentors.filter((mentor) => compared.includes(mentor.id))

  if (selectedMentor) {
    return (
      <StudentMentorProfile
        mentor={selectedMentor}
        isCompared={compared.includes(selectedMentor.id)}
        isSent={sentRequests.includes(selectedMentor.id)}
        onBack={() => setSelectedMentor(null)}
        onCompare={() => toggleCompare(selectedMentor.id)}
        onRequest={() => sendRequest(selectedMentor.id)}
      />
    )
  }

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
                      <button
                        type="button"
                        className={styles.profileLink}
                        onClick={() => setSelectedMentor(mentor)}
                        aria-label={`Открыть профиль ${mentor.name}`}
                      >
                        <Avatar name={mentor.name} size="l" />
                      </button>

                      <div className={styles.mentorInfo}>
                        <h3>
                          <h3>
                            <button
                              type="button"
                              className={styles.profileName}
                              onClick={() => setSelectedMentor(mentor)}
                            >
                              {mentor.name}
                            </button>
                          </h3>
                        </h3>
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
            <div className={styles.priceInputs}>
              <Input
                className={styles.priceInput}
                label="От, ₽"
                type="number"
                min={0}
                max={10000}
                value={String(priceMin)}
                onChange={(event) => {
                  const value = Math.min(10000, Math.max(0, Number(event.target.value) || 0))
                  setPriceMin(Math.min(value, priceMax))
                }}
              />
              <Input
                className={styles.priceInput}
                label="До, ₽"
                type="number"
                min={0}
                max={10000}
                value={String(priceMax)}
                onChange={(event) => {
                  const value = Math.min(10000, Math.max(0, Number(event.target.value) || 0))
                  setPriceMax(Math.max(value, priceMin))
                }}
              />
            </div>
            <p className={styles.filterHint}>Оплата — напрямую наставнику, вне платформы</p>
          </div>

          <div className={styles.filterGroup}>
            <Slider
              label="Время в неделю"
              value={hoursLimit}
              min={1}
              max={10}
              step={1}
              onChange={setHoursLimit}
              format={(value) => (value === 10 ? '10+ ч' : `${value} ч`)}
              marks={['1 ч', '5 ч', '10+ ч']}
            />
          </div>

          <div className={styles.filterGroup}>
            <Slider
              label="Минимальное совпадение"
              value={minMatch}
              min={0}
              max={100}
              step={5}
              onChange={setMinMatch}
              format={(value) => `от ${value}%`}
              marks={['0%', '50%', '100%']}
            />
          </div>

          <div className={styles.filterGroup}>
            <Slider
              label="Рейтинг"
              value={ratingMin}
              min={1}
              max={5}
              step={0.1}
              onChange={setRatingMin}
              format={(value) => `от ${value.toFixed(1).replace('.', ',')} ★`}
              marks={['1,0', '3,0', '5,0']}
            />
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
