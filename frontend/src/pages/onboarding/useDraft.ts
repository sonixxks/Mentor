import { useEffect, useState } from 'react'

export function useDraft<T extends object>(key: string, initial: T) {
  const [value, setValue] = useState<T>(() => {
    try {
      const saved = localStorage.getItem(key)
      return saved ? { ...initial, ...JSON.parse(saved) } : initial
    } catch {
      return initial
    }
  })

  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch {
    }
  }, [key, value])

  function update(patch: Partial<T>) {
    setValue((prev) => ({ ...prev, ...patch }))
  }

  function clear() {
    try {
      localStorage.removeItem(key)
    } catch {
    }
    setValue(initial)
  }

  return { value, update, clear }
}