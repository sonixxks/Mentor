import { StrictMode, useState } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './styles/tokens.css'
import './styles/global.css'
import { Button } from './components/ui/Button'
import { Input } from './components/ui/Input'
import { Tag } from './components/ui/Tag'
import { Badge } from './components/ui/Badge'
import { Card } from './components/ui/Card'
import { Avatar } from './components/ui/Avatar'
import { Logo } from './components/Logo/Logo'

function Showcase() {
  const [selected, setSelected] = useState(true)

  return (
    <main style={{ padding: 40, display: 'flex', flexDirection: 'column', gap: 32, maxWidth: 900 }}>
      <Logo />  

      <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap', alignItems: 'center' }}>
        <Button size="lg">Большая</Button>
        <Button>Primary</Button>
        <Button variant="accent">Accent</Button>
        <Button variant="secondary">Secondary</Button>
        <Button variant="danger">Danger</Button>
        <Button size="sm" variant="secondary">Маленькая</Button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
        <Input label="Почта" type="email" placeholder="ekaterina@urfu.me" />
        <Input label="Повторите пароль" type="password" error="Пароли не совпадают" />
      </div>

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        <Tag selected={selected} onClick={() => setSelected(!selected)}>
          Python
        </Tag>
        <Tag>Machine Learning</Tag>
        <Tag onRemove={() => alert('Удалить SQL')}>SQL</Tag>
      </div>

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <Badge variant="pending">Ожидает ответа</Badge>
        <Badge variant="review">На проверке</Badge>
        <Badge variant="approved">Одобрено</Badge>
        <Badge variant="rejected">Отклонено</Badge>
        <Badge variant="match">96% совпадение</Badge>
        <Badge>В работе</Badge>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 16 }}>
        <Card>
          <p className="t-body">Обычная карточка</p>
        </Card>
        <Card variant="accent">
          <p className="t-body">Акцентная</p>
        </Card>
        <Card variant="glass">
          <p className="t-body">Стеклянная</p>
        </Card>
      </div>

      <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
        <Avatar name="Екатерина Воробьева" size="s" />
        <Avatar name="Данил Гатаулин" />
        <Avatar name="Софья Власова" size="l" />
      </div>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Showcase />
    </BrowserRouter>
  </StrictMode>,
)