function splitOrderItemSegments(name: string | undefined) {
  return String(name || '')
    .split(' · ')
    .map((part) => part.trim())
    .filter(Boolean)
}

export function splitOrderItemName(name: string | undefined): [string, string] {
  const parts = splitOrderItemSegments(name)
  if (parts.length < 2) return ['', String(name || '')]
  return [parts[0], parts.slice(1).join(' · ')]
}

export function getReceiptItemDisplayName(name: string | undefined, isRetailGroup: boolean, retailPrefixes: string[] = []) {
  const rawName = String(name || '').trim()
  if (!rawName) return ''

  if (!isRetailGroup) {
    const [, itemName] = splitOrderItemName(rawName)
    return itemName
  }

  const parts = splitOrderItemSegments(rawName)
  if (parts.length <= 1) return rawName

  const normalizedPrefixes = new Set(
    ['零售商品', ...retailPrefixes]
      .map((part) => String(part || '').trim())
      .filter(Boolean),
  )

  let startIndex = 0
  while (startIndex < parts.length - 1 && normalizedPrefixes.has(parts[startIndex])) {
    startIndex += 1
  }
  if (startIndex > 0) {
    return parts.slice(startIndex).join(' · ')
  }

  // Backward-compatible fallback for legacy "宠物名 · 商品名(规格)" retail rows.
  const secondPart = parts[1]
  if (secondPart.includes('(') || secondPart.includes('（')) {
    return secondPart
  }

  return rawName
}
