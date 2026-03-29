const roleRank: Record<string, number> = {
  staff: 1,
  manager: 2,
  admin: 3,
}

export function normalizeStaffRole(role?: string | null): 'admin' | 'manager' | 'staff' {
  if (role === 'admin' || role === 'manager' || role === 'staff') {
    return role
  }
  return 'staff'
}

export function hasStaffRoleAtLeast(role: string | undefined | null, minRole: 'admin' | 'manager' | 'staff') {
  return roleRank[normalizeStaffRole(role)] >= roleRank[minRole]
}

export function staffRoleLabel(role?: string | null) {
  const normalized = normalizeStaffRole(role)
  if (normalized === 'admin') return '店长'
  if (normalized === 'manager') return '店员主管'
  return '店员'
}

export function compareStaffRole(aRole?: string | null, bRole?: string | null) {
  return roleRank[normalizeStaffRole(aRole)] - roleRank[normalizeStaffRole(bRole)]
}
