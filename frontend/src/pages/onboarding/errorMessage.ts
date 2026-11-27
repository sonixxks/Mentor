import { ApiError } from '../../api/client'

export function getSubmitErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 400) return `Сервер не принял анкету: ${error.message}`
    if (error.status === 401) return 'Сессия истекла. Войдите в аккаунт заново.'
    return error.message
  }
  return 'Не удалось отправить анкету. Попробуйте ещё раз.'
}