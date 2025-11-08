type ClassValue = string | number | false | null | undefined

export const cn = (...values: ClassValue[]): string => {
  return values
    .flatMap((value) => {
      if (!value && value !== 0) return []
      return String(value).split(' ')
    })
    .filter(Boolean)
    .join(' ')
}

