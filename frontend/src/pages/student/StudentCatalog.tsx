import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { clsx } from 'clsx'
import { Search, SlidersHorizontal } from 'lucide-react'
import { useState } from 'react'
import { getMentors } from '../../api/mentors'
import type { MentorFilters, MentorSort } from '../../api/mentors'
import { MentorCard } from '../../components/MentorCard'
import { RequestModal } from '../../components/RequestModal'
import type { RequestTarget } from '../../components/RequestModal'
import { Button } from '../../components/ui/Button'
import { Input } from '../../components/ui/Input'
import { Select } from '../../components/ui/Select'
import { Slider } from '../../components/ui/Slider'
import { Tag } from '../../components/ui/Tag'
import { useDebounce } from '../../hooks/useDebounce'
import { useSkills } from '../../hooks/useSkills'
import { plural } from '../../lib/plural'
import styles from './StudentCatalog.module.css'

const SORT_OPTIONS: { value: MentorSort; label: string }[] = [
  { value: 'match', label: 'По совместимости' },
  { value: 'rating', label: 'По рейтингу' },
  { value: 'price_asc', label: 'Сначала дешевле' },
  { value: 'price_desc', label: 'Сначала дороже' },
  { value: 'experience', label: 'По опыту' },
]

const PAGE_SIZE = 9
const SKILLS_VISIBLE = 6

const DEFAULTS = {
  query: '',
  sort: 'match' as MentorSort,
  skills: [] as string[],
  priceMin: '',
  priceMax: '',
  hoursMin: '',
  hoursMax: '',
  minMatch: 0,
  minRating: 0,
  onlyFree: true,
}

type FiltersState = typeof DEFAULTS

/** Пустое поле → undefined (фильтр не отправляем), иначе число */
const toNumber = (value: string) => (value.trim() === '' ? undefined : Number(value))

export function StudentCatalog() {
  const [state, setState] = useState<FiltersState>(DEFAULTS)
  const [limit, setLimit] = useState(PAGE_SIZE)
  const [showAllSkills, setShowAllSkills] = useState(false)
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [requestTo, setRequestTo] = useState<RequestTarget | null>(null)

  const query = useDebounce(state.query)
  const skillsQuery = useSkills()

  function update(patch: Partial<FiltersState>) {
    setState((prev) => ({ ...prev, ...patch }))
    setLimit(PAGE_SIZE)
  }

  function toggleSkill(skill: string) {
    update({
      skills: state.skills.includes(skill)
        ? state.skills.filter((item) => item !== skill)
        : [...state.skills, skill],
    })
  }

  const filters: MentorFilters = {
    q: query.trim() || undefined,
    skills: state.skills,
    sort: state.sort,
    price_min: toNumber(state.priceMin),
    price_max: toNumber(state.priceMax),
    hours_min: toNumber(state.hoursMin),
    hours_max: toNumber(state.hoursMax),
    min_match: state.minMatch || undefined,
    min_rating: state.minRating || undefined,
    has_free_slots: state.onlyFree,
    limit,
  }

  const mentors = useQuery({
    queryKey: ['mentors', filters],
    queryFn: () => getMentors(filters),
    placeholderData: keepPreviousData,
  })

  const total = mentors.data?.total ?? 0
  const items = mentors.data?.items ?? []

  const allSkills = skillsQuery.data ?? []
  const visibleSkills = showAllSkills ? allSkills : allSkills.slice(0, SKILLS_VISIBLE)
  // Выбранные навыки видны всегда, даже если список свёрнут
  const shownSkills = [...new Set([...state.skills, ...visibleSkills])]

  return (
    <div className={styles.layout}>
      <section className={styles.results}>
        <div className={styles.toolbar}>
          <div className={styles.search}>
            <Search size={20} className={styles.searchIcon} />
            <Input
              placeholder="Навык, имя или компания"
              value={state.query}
              onChange={(e) => update({ query: e.target.value })}
            />
          </div>
          <div className={styles.sort}>
            <Select
              options={SORT_OPTIONS}
              value={state.sort}
              onChange={(e) => update({ sort: e.target.value as MentorSort })}
            />
          </div>
          <Button
            variant="secondary"
            size="md"
            className={styles.filtersToggle}
            onClick={() => setFiltersOpen((open) => !open)}
          >
            <SlidersHorizontal size={16} />
            Фильтры
          </Button>
        </div>

        {mentors.isPending ? <p className={styles.count}>Загружаем наставников…</p> : null}

        {mentors.isError ? (
          <div className={styles.state}>
            <p>Не удалось загрузить наставников.</p>
            <Button variant="secondary" size="md" onClick={() => mentors.refetch()}>
              Попробовать ещё раз
            </Button>
          </div>
        ) : null}

        {mentors.isSuccess ? (
          <>
            <p className={styles.count}>
              Найдено {plural(total, ['наставник', 'наставника', 'наставников'])}
            </p>

            {items.length > 0 ? (
              <div className={clsx(styles.grid, mentors.isPlaceholderData ? styles.fetching : '')}>
                {items.map((mentor) => (
                  <MentorCard
                    key={mentor.id}
                    mentor={mentor}
                    onRequest={() => setRequestTo(mentor)}
                  />
                ))}
              </div>
            ) : (
              <div className={styles.state}>
                <p>По этим фильтрам наставников нет. Попробуй смягчить условия.</p>
                <Button variant="secondary" size="md" onClick={() => update(DEFAULTS)}>
                  Сбросить фильтры
                </Button>
              </div>
            )}

            {items.length < total ? (
              <Button
                variant="secondary"
                size="md"
                className={styles.more}
                onClick={() => setLimit((value) => value + PAGE_SIZE)}
              >
                Показать ещё
              </Button>
            ) : null}
          </>
        ) : null}
      </section>

      <aside className={clsx(styles.filters, filtersOpen ? styles.filtersOpen : '')}>
        <div className={styles.filtersHead}>
          <h2 className={styles.filtersTitle}>Фильтры</h2>
          <button type="button" className={styles.reset} onClick={() => update(DEFAULTS)}>
            Сбросить
          </button>
        </div>

        <div className={styles.group}>
          <span className={styles.groupTitle}>Навыки</span>
          <div className={styles.tags}>
            {shownSkills.map((skill) => (
              <Tag
                key={skill}
                selected={state.skills.includes(skill)}
                onClick={() => toggleSkill(skill)}
              >
                {skill}
              </Tag>
            ))}
            {allSkills.length > SKILLS_VISIBLE ? (
              <button
                type="button"
                className={styles.moreSkills}
                onClick={() => setShowAllSkills((value) => !value)}
              >
                {showAllSkills ? 'Свернуть' : '+ ещё'}
              </button>
            ) : null}
          </div>
        </div>

        <div className={styles.group}>
          <span className={styles.groupTitle}>Стоимость занятия, ₽</span>
          <div className={styles.pair}>
            <Input
              type="number"
              min={0}
              placeholder="от"
              value={state.priceMin}
              onChange={(e) => update({ priceMin: e.target.value })}
            />
            <Input
              type="number"
              min={0}
              placeholder="до"
              value={state.priceMax}
              onChange={(e) => update({ priceMax: e.target.value })}
            />
          </div>
          <p className={styles.hint}>0 — бесплатно. Оплата напрямую наставнику, вне платформы</p>
        </div>

        <div className={styles.group}>
          <span className={styles.groupTitle}>Время в неделю, ч</span>
          <div className={styles.pair}>
            <Input
              type="number"
              min={1}
              max={10}
              placeholder="от"
              value={state.hoursMin}
              onChange={(e) => update({ hoursMin: e.target.value })}
            />
            <Input
              type="number"
              min={1}
              max={10}
              placeholder="до"
              value={state.hoursMax}
              onChange={(e) => update({ hoursMax: e.target.value })}
            />
          </div>
        </div>

        <Slider
          label="Минимальное совпадение"
          min={0}
          max={100}
          step={5}
          value={state.minMatch}
          onChange={(minMatch) => update({ minMatch })}
          format={(value) => (value === 0 ? 'любое' : `от ${value}%`)}
        />

        <Slider
          label="Рейтинг"
          min={0}
          max={5}
          step={0.5}
          value={state.minRating}
          onChange={(minRating) => update({ minRating })}
          format={(value) => (value === 0 ? 'любой' : `от ${String(value).replace('.', ',')}`)}
        />

        <label className={styles.switchRow}>
          <span>Есть свободные места</span>
          <input
            type="checkbox"
            className={styles.switchInput}
            checked={state.onlyFree}
            onChange={(e) => update({ onlyFree: e.target.checked })}
          />
          <span className={styles.switch} aria-hidden="true" />
        </label>
      </aside>

      {requestTo ? <RequestModal mentor={requestTo} onClose={() => setRequestTo(null)} /> : null}
    </div>
  )
}
