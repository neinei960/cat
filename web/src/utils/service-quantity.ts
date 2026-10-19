export function serviceQuantity(value: unknown): number {
  const quantity = Number(value)
  return Number.isInteger(quantity) && quantity > 0 ? quantity : 1
}

export function canRepeatService(name: string): boolean {
  return !name.includes('全年口腔护理')
}

export function serviceSelectionKey(serviceId: number, specId: number, name: string, price: number): string {
  return JSON.stringify([serviceId, specId || 0, name.trim(), Number(price).toFixed(2)])
}

export function serviceLineAmount(price: number, quantity: unknown): number {
  return Math.round(price * 100 * serviceQuantity(quantity)) / 100
}
