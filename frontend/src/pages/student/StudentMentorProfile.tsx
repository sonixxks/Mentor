import { ArrowLeft, CheckCircle2, Clock3, Star } from 'lucide-react'
import { Avatar } from '../../components/ui/Avatar/Avatar'
import { Button } from '../../components/ui/Button/Button'
import { Badge } from '../../components/ui/Badge/Badge'
import styles from './StudentMentorProfile.module.css'

export type Mentor = {
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

type Props = {
  mentor: Mentor
  isCompared: boolean
  isSent: boolean
  onBack: () => void
  onCompare: () => void
  onRequest: () => void
}

const details: Record<
  number,
  {
    about: string
    skills: string[]
    goals: number
    freeTime: number
    communication: number
    programs: {
      title: string
      description: string
      lessons: string
      price: string
      result: string
    }[]
    reviews: { text: string; author: string }[]
  }
> = {
  1: {
    about:
      'Семь лет работаю с нейросетями: от прототипов до продакшена. Помогаю студентам собрать первый проект для портфолио и подготовиться к стажировке. Люблю чёткий план и короткие созвоны раз в неделю.',
    skills: [
      'Python',
      'PyTorch',
      'Нейросети',
      'Computer Vision',
      'MLOps',
      'Подготовка к собеседованию',
    ],
    goals: 92,
    freeTime: 85,
    communication: 90,
    programs: [
      {
        title: 'Первый проект в Computer Vision',
        description:
          'Датасет и разметка, обучение модели детекции, оценка качества, упаковка в портфолио.',
        lessons: '12 занятий',
        price: '1 250 ₽ за занятие',
        result: 'Проект на GitHub с рабочей моделью и описанием.',
      },
      {
        title: 'Подготовка к стажировке в DS',
        description: 'Резюме и портфолио, разбор типовых задач, два пробных собеседования.',
        lessons: '6 занятий',
        price: 'Бесплатно',
        result: 'Готовое резюме и уверенность на собеседовании.',
      },
    ],
    reviews: [
      {
        text: 'За три месяца собрала проект по компьютерному зрению и прошла на стажировку.',
        author: 'Анна, 4 курс · сентябрь 2026',
      },
      {
        text: 'Объясняет сложное простыми словами и всегда на связи.',
        author: 'Илья, выпускник · июль 2026',
      },
    ],
  },
}

const fallback = {
  about:
    'Помогаю ученикам развивать практические навыки, разбираться в сложных темах и двигаться к профессиональным целям. На занятиях уделяю внимание практике и понятному плану обучения.',
  skills: ['Практические проекты', 'Развитие навыков', 'Подготовка к собеседованию'],
  goals: 92,
  freeTime: 85,
  communication: 90,
  programs: [
    {
      title: 'Индивидуальная программа обучения',
      description: 'План занятий с практическими задачами и обратной связью.',
      lessons: '8 занятий',
      price: '1 200 ₽ за занятие',
      result: 'План развития и готовый учебный проект.',
    },
  ],
  reviews: [
    {
      text: 'Занятия помогли структурировать знания и увереннее двигаться дальше.',
      author: 'Ученик · 2026',
    },
  ],
}

const priceLabel = (price: number) =>
  price === 0 ? 'Бесплатно' : `${price.toLocaleString('ru-RU')} ₽ за занятие`

export function StudentMentorProfile({
  mentor,
  isCompared,
  isSent,
  onBack,
  onCompare,
  onRequest,
}: Props) {
  const content = details[mentor.id] ?? fallback

  return (
    <main className={styles.profile}>
      <header className={styles.heading}>
        <button type="button" className={styles.back} onClick={onBack}>
          <ArrowLeft size={20} />
          Назад
        </button>
        <h1>Профиль наставника</h1>
      </header>

      <div className={styles.columns}>
        <aside className={styles.leftColumn}>
          <section className={styles.identityCard}>
            <div className={styles.portrait}>
              <Avatar name={mentor.name} size="l" />
            </div>

            <span
              className={`${styles.availability} ${mentor.freePlaces ? '' : styles.unavailable}`}
            >
              <i />
              {mentor.freePlaces ? 'Есть свободное место' : 'Нет свободных мест'}
            </span>

            <h2>{mentor.name}</h2>
            <p className={styles.position}>
              {mentor.position} · {mentor.experience}
            </p>

            <Badge variant="approved">Опыт подтверждён куратором</Badge>

            <div className={styles.rating}>
              <Star size={16} fill="currentColor" />
              <strong>{mentor.rating.toFixed(1).replace('.', ',')}</strong>
              <span>{mentor.reviews} отзывов</span>
            </div>

            <Button fullWidth size="lg" onClick={onRequest} disabled={isSent}>
              {isSent ? 'Заявка отправлена' : 'Отправить заявку'}
            </Button>

            <Button fullWidth variant="secondary" size="lg" onClick={onCompare}>
              {isCompared ? 'Убрать из сравнения' : 'Добавить к сравнению'}
            </Button>
          </section>

          <section className={styles.availabilityCard}>
            <h3>Доступность</h3>
            <dl>
              <div>
                <dt>Время</dt>
                <dd>{Math.round(mentor.hours)} ч в неделю</dd>
              </div>
              <div>
                <dt>Когда</dt>
                <dd>{mentor.schedule}</dd>
              </div>
              <div>
                <dt>Срок</dt>
                <dd>от 1 до 6 месяцев</dd>
              </div>
              <div>
                <dt>Цена</dt>
                <dd>{priceLabel(mentor.price)}</dd>
              </div>
            </dl>
            <p>Оплата — напрямую наставнику, платформа не участвует в расчётах</p>
          </section>
        </aside>

        <section className={styles.rightColumn}>
          <section className={styles.matchCard}>
            <div className={styles.matchScore}>
              <strong>{mentor.match}%</strong>
              <span>совместимость с тобой</span>
            </div>

            <div className={styles.metrics}>
              <Metric label="Навыки и интересы" value={96} />
              <Metric label="Цели" value={content.goals} />
              <Metric label="Свободное время" value={content.freeTime} />
              <Metric label="Стиль общения" value={content.communication} />
            </div>
          </section>

          <section className={styles.aboutCard}>
            <h3>О себе</h3>
            <p>{content.about}</p>
            <div className={styles.skills}>
              {content.skills.map((skill) => (
                <span key={skill}>{skill}</span>
              ))}
            </div>
          </section>

          <div className={styles.lowerGrid}>
            <section className={styles.programsCard}>
              <div className={styles.sectionHeading}>
                <h3>Авторские программы</h3>
                <span>или открытый запрос — план составите вместе</span>
              </div>

              {content.programs.map((program) => (
                <article className={styles.program} key={program.title}>
                  <div className={styles.programTop}>
                    <h4>{program.title}</h4>
                    <strong>
                      {program.lessons} · {program.price}
                    </strong>
                  </div>
                  <p>{program.description}</p>
                  <p className={styles.outcome}>
                    <CheckCircle2 size={15} />
                    Результат: {program.result}
                  </p>
                </article>
              ))}
            </section>

            <section className={styles.reviewsCard}>
              <div className={styles.sectionHeading}>
                <h3>Отзывы учеников</h3>
                <span>Все {mentor.reviews}</span>
              </div>

              {content.reviews.map((review) => (
                <article className={styles.review} key={review.author}>
                  <div className={styles.stars}>★★★★★</div>
                  <p>«{review.text}»</p>
                  <span>{review.author}</span>
                </article>
              ))}
            </section>
          </div>

          <p className={styles.demoNote}>
            <Clock3 size={13} />
            Демонстрационный профиль. Информация и отзывы пока представлены примерами.
          </p>
        </section>
      </div>
    </main>
  )
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div className={styles.metric}>
      <div>
        <span>{label}</span>
        <strong>{value}%</strong>
      </div>
      <i>
        <b style={{ width: `${value}%` }} />
      </i>
    </div>
  )
}
