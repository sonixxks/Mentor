import { ApiError } from '../api/client'

const ERROR_TEXT: Record<string, string> = {
  VALIDATION_ERROR: 'Проверь, что все поля заполнены правильно.',
  UNAUTHORIZED: 'Сессия истекла. Войди в аккаунт заново.',
  INVALID_TOKEN: 'Сессия истекла. Войди в аккаунт заново.',
  FORBIDDEN: 'Для этого действия не хватает прав.',
  ROLE_REQUIRED: 'Для этого действия не хватает прав.',
  NOT_FOUND: 'Не нашли нужные данные. Возможно, их уже удалили.',
  CONFLICT: 'Данные уже изменились. Обнови страницу.',
  INVALID_CODE: 'Неверный код. Проверь письмо и попробуй ещё раз.',
  TOO_MANY_REQUESTS: 'Слишком много попыток. Подожди немного и попробуй снова.',
  PROFILE_INCOMPLETE: 'Сначала заполни анкету до конца.',
  MENTOR_LIMIT_REACHED: 'У наставника закончились свободные места. Выбери другого наставника.',
  REQUEST_LIMIT_REACHED:
    'У тебя уже 3 активные заявки — это максимум. Дождись ответа или отмени одну в «Моих заявках».',
  DUPLICATE_REQUEST: 'Заявка этому наставнику уже отправлена. Её статус — в «Моих заявках».',
  REVIEW_EXISTS: 'Отзыв об этой встрече уже оставлен.',
  INVALID_TRANSITION: 'Статус заявки уже изменился. Обнови страницу.',
}

export function errorMessage(error: unknown, fallback = 'Не получилось. Попробуй ещё раз.') {
  if (error instanceof ApiError) {
    if (error.code && ERROR_TEXT[error.code]) return ERROR_TEXT[error.code]
    if (error.status >= 500) return 'Сервер сейчас не отвечает. Попробуй чуть позже.'
  }
  if (error instanceof TypeError) return 'Нет связи с сервером. Проверь интернет.'
  return fallback
}
