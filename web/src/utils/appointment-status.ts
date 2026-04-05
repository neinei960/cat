export interface AppointmentStatusMeta {
  label: string
  badgeBg: string
  badgeText: string
  badgeBorder: string
  barBg: string
  barText: string
  barShadow: string
  blockBg: string
  blockAccent: string
  blockBorder: string
  blockShadow: string
  opacity?: string
}

const DEFAULT_STATUS_META: AppointmentStatusMeta = {
  label: '未知状态',
  badgeBg: '#F3F4F6',
  badgeText: '#4B5563',
  badgeBorder: 'rgba(107, 114, 128, 0.18)',
  barBg: 'linear-gradient(135deg, #F8FAFC 0%, #E5E7EB 100%)',
  barText: '#475569',
  barShadow: '0 12rpx 24rpx rgba(148, 163, 184, 0.12)',
  blockBg: 'linear-gradient(180deg, #F8FAFC 0%, #E5E7EB 100%)',
  blockAccent: '#94A3B8',
  blockBorder: 'rgba(148, 163, 184, 0.18)',
  blockShadow: '0 10rpx 18rpx rgba(148, 163, 184, 0.08)',
}

export const APPOINTMENT_STATUS_META: Record<number, AppointmentStatusMeta> = {
  0: {
    label: '待确认',
    badgeBg: '#FEF3C7',
    badgeText: '#92400E',
    badgeBorder: 'rgba(217, 119, 6, 0.18)',
    barBg: 'linear-gradient(135deg, #FFF7ED 0%, #FDE68A 100%)',
    barText: '#92400E',
    barShadow: '0 14rpx 28rpx rgba(217, 119, 6, 0.16)',
    blockBg: 'linear-gradient(180deg, #FFF8E7 0%, #FDE7AF 100%)',
    blockAccent: '#D97706',
    blockBorder: 'rgba(217, 119, 6, 0.20)',
    blockShadow: '0 12rpx 24rpx rgba(217, 119, 6, 0.12)',
  },
  1: {
    label: '已确认',
    badgeBg: '#E0E7FF',
    badgeText: '#3730A3',
    badgeBorder: 'rgba(79, 70, 229, 0.18)',
    barBg: 'linear-gradient(135deg, #EEF2FF 0%, #C7D2FE 100%)',
    barText: '#3730A3',
    barShadow: '0 14rpx 28rpx rgba(79, 70, 229, 0.14)',
    blockBg: 'linear-gradient(180deg, #F5F7FF 0%, #DEE8FF 100%)',
    blockAccent: '#4338CA',
    blockBorder: 'rgba(67, 56, 202, 0.20)',
    blockShadow: '0 12rpx 24rpx rgba(67, 56, 202, 0.12)',
  },
  2: {
    label: '服务中',
    badgeBg: '#DCFCE7',
    badgeText: '#166534',
    badgeBorder: 'rgba(5, 150, 105, 0.18)',
    barBg: 'linear-gradient(135deg, #ECFDF5 0%, #BBF7D0 100%)',
    barText: '#166534',
    barShadow: '0 14rpx 28rpx rgba(5, 150, 105, 0.14)',
    blockBg: 'linear-gradient(180deg, #F0FDF4 0%, #CFFCE2 100%)',
    blockAccent: '#059669',
    blockBorder: 'rgba(5, 150, 105, 0.18)',
    blockShadow: '0 12rpx 24rpx rgba(5, 150, 105, 0.12)',
  },
  3: {
    label: '待结算',
    badgeBg: '#E0F2FE',
    badgeText: '#075985',
    badgeBorder: 'rgba(2, 132, 199, 0.18)',
    barBg: 'linear-gradient(135deg, #F0F9FF 0%, #BAE6FD 100%)',
    barText: '#075985',
    barShadow: '0 14rpx 28rpx rgba(2, 132, 199, 0.14)',
    blockBg: 'linear-gradient(180deg, #F0F9FF 0%, #CDEEFF 100%)',
    blockAccent: '#0284C7',
    blockBorder: 'rgba(2, 132, 199, 0.18)',
    blockShadow: '0 10rpx 18rpx rgba(2, 132, 199, 0.10)',
  },
  7: {
    label: '已开单',
    badgeBg: '#EDE9FE',
    badgeText: '#6D28D9',
    badgeBorder: 'rgba(109, 40, 217, 0.18)',
    barBg: 'linear-gradient(135deg, #F5F3FF 0%, #DDD6FE 100%)',
    barText: '#6D28D9',
    barShadow: '0 14rpx 28rpx rgba(109, 40, 217, 0.14)',
    blockBg: 'linear-gradient(180deg, #F5F3FF 0%, #E9DDFF 100%)',
    blockAccent: '#7C3AED',
    blockBorder: 'rgba(124, 58, 237, 0.18)',
    blockShadow: '0 10rpx 18rpx rgba(124, 58, 237, 0.10)',
  },
  4: {
    label: '已取消',
    badgeBg: '#E5E7EB',
    badgeText: '#4B5563',
    badgeBorder: 'rgba(107, 114, 128, 0.18)',
    barBg: 'linear-gradient(135deg, #F8FAFC 0%, #E5E7EB 100%)',
    barText: '#4B5563',
    barShadow: '0 10rpx 20rpx rgba(107, 114, 128, 0.10)',
    blockBg: 'linear-gradient(180deg, #F8FAFC 0%, #E5E7EB 100%)',
    blockAccent: '#6B7280',
    blockBorder: 'rgba(107, 114, 128, 0.16)',
    blockShadow: 'none',
    opacity: '0.72',
  },
  6: {
    label: '已到店',
    badgeBg: '#FDF4FF',
    badgeText: '#86198F',
    badgeBorder: 'rgba(168, 85, 247, 0.18)',
    barBg: 'linear-gradient(135deg, #FAF5FF 0%, #E9D5FF 100%)',
    barText: '#86198F',
    barShadow: '0 14rpx 28rpx rgba(168, 85, 247, 0.14)',
    blockBg: 'linear-gradient(180deg, #FAF5FF 0%, #F0E0FF 100%)',
    blockAccent: '#A855F7',
    blockBorder: 'rgba(168, 85, 247, 0.20)',
    blockShadow: '0 12rpx 24rpx rgba(168, 85, 247, 0.12)',
  },
  5: {
    label: '未到店',
    badgeBg: '#FEE2E2',
    badgeText: '#B91C1C',
    badgeBorder: 'rgba(220, 38, 38, 0.18)',
    barBg: 'linear-gradient(135deg, #FEF2F2 0%, #FECACA 100%)',
    barText: '#B91C1C',
    barShadow: '0 14rpx 28rpx rgba(220, 38, 38, 0.16)',
    blockBg: 'linear-gradient(180deg, #FEF2F2 0%, #FECACA 100%)',
    blockAccent: '#DC2626',
    blockBorder: 'rgba(220, 38, 38, 0.18)',
    blockShadow: '0 10rpx 18rpx rgba(220, 38, 38, 0.10)',
    opacity: '0.84',
  },
}

export function getAppointmentStatusMeta(status?: number): AppointmentStatusMeta {
  if (typeof status !== 'number') return DEFAULT_STATUS_META
  if (status === 6) return APPOINTMENT_STATUS_META[1]
  return APPOINTMENT_STATUS_META[status] || DEFAULT_STATUS_META
}

export function getAppointmentStatusLabel(status?: number): string {
  return getAppointmentStatusMeta(status).label
}

export function getAppointmentStatusBadgeStyle(status?: number) {
  const meta = getAppointmentStatusMeta(status)
  return {
    background: meta.badgeBg,
    color: meta.badgeText,
    border: `1rpx solid ${meta.badgeBorder}`,
  }
}

export function getAppointmentStatusBarStyle(status?: number) {
  const meta = getAppointmentStatusMeta(status)
  return {
    background: meta.barBg,
    color: meta.barText,
    boxShadow: meta.barShadow,
  }
}

export function getAppointmentStatusBlockStyle(status?: number) {
  const meta = getAppointmentStatusMeta(status)
  return {
    borderLeftColor: meta.blockAccent,
    background: meta.blockBg,
    borderColor: meta.blockBorder,
    boxShadow: meta.blockShadow,
    opacity: meta.opacity || '1',
  }
}

export function getAppointmentStatusTabStyle(status: number, active: boolean) {
  if (!active) return {}
  const meta = getAppointmentStatusMeta(status)
  return {
    background: meta.badgeBg,
    color: meta.badgeText,
    borderColor: meta.badgeBorder,
  }
}
