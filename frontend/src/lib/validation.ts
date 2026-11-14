export const PASSWORD_MIN = 8

export function isEmail(value: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim())
}

export function isUrl(value: string) {
  return /^(https?:\/\/)?[^\s.]+\.[^\s]{2,}$/.test(value)
}