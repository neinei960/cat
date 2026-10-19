type ServiceCategoryLike = {
  ID: number
  name?: string
  children?: ServiceCategoryLike[]
}

export const HIDDEN_SERVICE_MANAGEMENT_CATEGORY_NAMES = new Set(['寄养托管', '上门喂养'])

function normalizeCategoryName(name?: string) {
  return String(name || '').trim()
}

function isHiddenServiceManagementRoot(category: ServiceCategoryLike) {
  return HIDDEN_SERVICE_MANAGEMENT_CATEGORY_NAMES.has(normalizeCategoryName(category.name))
}

export function filterServiceManagementCategoryTree<T extends ServiceCategoryLike>(categories: T[]): T[] {
  return categories.filter((category) => !isHiddenServiceManagementRoot(category))
}

export function getHiddenServiceManagementCategoryIds(categories: ServiceCategoryLike[]): Set<number> {
  const ids = new Set<number>()

  function collect(category: ServiceCategoryLike) {
    ids.add(category.ID)
    ;(category.children || []).forEach(collect)
  }

  categories.forEach((category) => {
    if (isHiddenServiceManagementRoot(category)) {
      collect(category)
    }
  })

  return ids
}
