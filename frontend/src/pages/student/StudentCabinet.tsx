import { useState } from 'react'
import { useLocation } from 'react-router-dom'
import {
  Bell,
  CalendarDays,
  Check,
  ChevronRight,
  LayoutGrid,
  LogOut,
  MessageCircle,
  Search,
  Sparkles,
  BriefcaseBusiness,
  SlidersHorizontal,
  CircleHelp,
} from 'lucide-react'
import { Logo } from '../../components/Logo/Logo'
import { Avatar } from '../../components/ui/Avatar/Avatar'
import { Badge } from '../../components/ui/Badge/Badge'
import { Button } from '../../components/ui/Button/Button'
import { Card } from '../../components/ui/Card/Card'
import { useSession } from '../../hooks/useSession'
import styles from './StudentCabinet.module.css'

type Mentor = {
  id: number
  name: string
  position: string
  experience: string
  rating: string
  match: number
  skills: string[]
  hours: string
  price: string
}

const mentors: Mentor[] = [
  {
    id: 1,
    name: 'Данил Королёв',
    position: 'ML-инженер',
    experience: '5 лет',
    rating: '4,9',
    match: 94,
    skills: ['Python', 'ML'],
    hours: '2 ч/нед · Вт, Чт 19–21',
    price: '1 000 ₽/занятие',
  },
  {
    id: 2,
    name: 'Мария Петрова',
    position: 'Data Analyst',
    experience: '6 лет',
    rating: '4,8',
    match: 91,
    skills: ['SQL', 'Статистика'],
    hours: '2 ч/нед · Сб 10–14',
    price: '1 200 ₽/занятие',
  },
  {
    id: 3,
    name: 'Виктор Соколов',
    position: 'CV-инженер',
    experience: '8 лет',
    rating: '4,7',
    match: 87,
    skills: ['Computer Vision'],
    hours: '1 ч/нед · Ср 20–21',
    price: '1 500 ₽/занятие',
  },
  {
    id: 4,
    name: 'Ольга Лебедева',
    position: 'Тимлид DS',
    experience: '10 лет',
    rating: '5,0',
    match: 84,
    skills: ['Карьера', 'MLOps'],
    hours: '2 ч/нед · Пн, Пт 8–9',
    price: 'Бесплатно',
  },
  {
    id: 5,
    name: 'Игорь Новиков',
    position: 'Backend + ML',
    experience: '4 года',
    rating: '4,6',
    match: 79,
    skills: ['Python', 'API'],
    hours: '3 ч/нед · Вт–Чт',
    price: '900 ₽/занятие',
  },
]

export function StudentCabinet() {
  const { logout } = useSession()
  const location = useLocation()

  const isEmptyDemo =
    import.meta.env.DEV && new URLSearchParams(location.search).get('empty') === '1'

  const [sentRequests, setSentRequests] = useState<number[]>([])
  const [visibleCount, setVisibleCount] = useState(4)

  const sendRequest = (id: number) => {
    setSentRequests((current) => (current.includes(id) ? current : [...current, id]))
  }

  return (
    <div className={styles.shell}>
      <aside className={styles.sidebar}>
        <div className={styles.logo}>
          <Logo to="/student" />
        </div>

        <nav className={styles.navigation}>
          <div className={`${styles.navItem} ${styles.navActive}`}>
            <Sparkles size={19} />
            <span>Подбор</span>
          </div>
          <div className={styles.navItem}>
            <Search size={19} />
            <span>Каталог</span>
          </div>
          <div className={styles.navItem}>
            <BriefcaseBusiness size={19} />
            <span>Мои заявки</span>
          </div>
          <div className={styles.navItem}>
            <LayoutGrid size={19} />
            <span>Пространство</span>
          </div>
          <div className={styles.navItem}>
            <CalendarDays size={19} />
            <span>Календарь</span>
          </div>
          <div className={styles.navItem}>
            <MessageCircle size={19} />
            <span>Чат</span>
            <span className={styles.chatCount}>2</span>
          </div>
        </nav>

        <Card className={styles.helpCard}>
          <h3>Не можешь выбрать?</h3>
          <p>Куратор поможет подобрать наставника вручную.</p>
          <button type="button" className={styles.textButton}>
            Запросить помощь
          </button>
        </Card>

        <button type="button" className={styles.logout} onClick={logout}>
          <LogOut size={19} />
          <span>Выйти</span>
        </button>
      </aside>

      <main className={styles.main}>
        <header className={styles.topbar}>
          <h1>Подбор</h1>
          <div className={styles.userInfo}>
            <button type="button" className={styles.iconButton} aria-label="Уведомления">
              <Bell size={20} />
            </button>
            <Avatar name="Екатерина Смирнова" size="m" />
          </div>
        </header>

        {isEmptyDemo ? (
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
                По Computer Vision и твоему расписанию совпадений выше 60% пока нет. Новые
                наставники регистрируются каждую неделю — сообщим, как только появится подходящий.
              </p>

              <div className={styles.emptyActions}>
                <button type="button" className={styles.outlineButton}>
                  <CircleHelp size={18} />
                  Запросить помощь куратора
                </button>

                <button type="button" className={styles.outlineButton}>
                  Смягчить фильтры в каталоге
                </button>
              </div>

              <button type="button" className={styles.refineButton}>
                Уточнить анкету
              </button>
            </section>
          </div>
        ) : (
          <div className={styles.content}>
            <section className={styles.heading}>
              <div>
                <h2>10 наставников под твои цели</h2>
                <p>Data Science, нейросети · 2 ч в неделю · стажировка за 6 месяцев</p>
              </div>
              <div className={styles.headingActions}>
                <Button variant="secondary" size="md" disabled>
                  Уточнить анкету
                </Button>
                <Button variant="secondary" size="md" disabled>
                  <SlidersHorizontal size={16} />
                  Каталог с фильтрами
                </Button>
              </div>
            </section>

            <section className={styles.featured}>
              <div className={styles.featuredAvatar}>
                <Avatar name="Алексей Миронов" size="l" />
              </div>

              <div className={styles.featuredInfo}>
                <Badge variant="match">Лучшее совпадение</Badge>
                <h2>Алексей Миронов</h2>
                <p>Senior Data Scientist · 7 лет в нейросетях · рейтинг 4,9 · 2 ч в неделю</p>
                <div className={styles.featuredActions}>
                  <Button onClick={() => sendRequest(100)} disabled={sentRequests.includes(100)}>
                    {sentRequests.includes(100) ? (
                      <>
                        <Check size={16} /> Заявка отправлена
                      </>
                    ) : (
                      'Отправить заявку'
                    )}
                  </Button>
                  <Button variant="secondary" disabled>
                    Профиль
                  </Button>
                </div>
              </div>

              <div className={styles.matchPanel}>
                <div className={styles.matchValue}>
                  <strong>99%</strong>
                  <span>совместимость</span>
                </div>
                {[
                  ['Навыки', 96],
                  ['Цели', 92],
                  ['Время', 85],
                  ['Стиль', 90],
                ].map(([label, value]) => (
                  <div className={styles.matchRow} key={label}>
                    <span>{label}</span>
                    <div className={styles.progress}>
                      <span style={{ width: `${value}%` }} />
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className={styles.mentorGrid}>
              {mentors.slice(0, visibleCount).map((mentor, index) => (
                <Card className={styles.mentorCard} key={mentor.id}>
                  <div className={styles.cardVisual}>
                    <Badge variant="match">{mentor.match}% совпадение</Badge>
                    <span>№{index + 2}</span>
                    <Avatar name={mentor.name} size="l" />
                  </div>

                  <div className={styles.mentorName}>
                    <h3>{mentor.name}</h3>
                    <span>★ {mentor.rating}</span>
                  </div>
                  <p className={styles.position}>
                    {mentor.position} · {mentor.experience}
                  </p>

                  <div className={styles.skills}>
                    {mentor.skills.map((skill) => (
                      <span key={skill}>{skill}</span>
                    ))}
                  </div>

                  <div className={styles.mentorMeta}>
                    <span>
                      <CalendarDays size={15} />
                      {mentor.hours}
                    </span>
                    <strong>{mentor.price}</strong>
                  </div>

                  <div className={styles.cardActions}>
                    <Button
                      size="sm"
                      fullWidth
                      onClick={() => sendRequest(mentor.id)}
                      disabled={sentRequests.includes(mentor.id)}
                    >
                      {sentRequests.includes(mentor.id) ? 'Заявка отправлена' : 'Отправить заявку'}
                    </Button>
                    <button
                      type="button"
                      className={styles.compareButton}
                      aria-label={`Добавить ${mentor.name} к сравнению`}
                      disabled
                    >
                      +
                    </button>
                  </div>
                </Card>
              ))}
            </section>

            {visibleCount < mentors.length && (
              <button
                type="button"
                className={styles.showMore}
                onClick={() => setVisibleCount(mentors.length)}
              >
                Показать ещё 5 из подборки <ChevronRight size={16} />
              </button>
            )}

            <p className={styles.demoNote}>
              Демонстрационные данные. Заявки пока не отправляются на сервер.
            </p>
          </div>
        )}
      </main>
    </div>
  )
}
