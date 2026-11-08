export function plural(n: number, forms: [string, string, string]) {
  const last = n % 10
  const lastTwo = n % 100

  if (last === 1 && lastTwo !== 11) return `${n} ${forms[0]}`
  if (last >= 2 && last <= 4 && (lastTwo < 12 || lastTwo > 14)) return `${n} ${forms[1]}`
  return `${n} ${forms[2]}`
}