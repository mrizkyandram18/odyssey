export type PaceStatus = 'ACHIEVED' | 'CRITICAL' | 'BEHIND' | 'ON_TRACK'

export interface TargetPaceInfo {
  target: number
  earned: number
  targetPercent: number
  monthProgress: number
  dayOfMonth: number
  daysInMonth: number
  status: PaceStatus
  label: string
  badgeClass: string
  barGradient: string
  message: string
}

/**
 * Calculates monthly target checkpoint pace and performance status.
 * - ACHIEVED: earned >= target (100%+)
 * - CRITICAL: dayOfMonth >= 20 and targetPercent < 50%
 * - BEHIND: dayOfMonth >= 12 and targetPercent < 60% of monthProgress
 * - ON_TRACK: pace is healthy
 */
export function calculateTargetPace(
  earned: number,
  target: number,
  currentDate = new Date()
): TargetPaceInfo {
  const dayOfMonth = currentDate.getDate()
  const daysInMonth = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 0).getDate()
  const monthProgress = Math.min(100, Math.max(1, Math.round((dayOfMonth / daysInMonth) * 100)))
  const safeTarget = Math.max(0, target)
  const targetPercent = safeTarget > 0 ? Math.min(100, Math.round((earned / safeTarget) * 100)) : 0

  if (safeTarget <= 0) {
    return {
      target: safeTarget,
      earned,
      targetPercent: 0,
      monthProgress,
      dayOfMonth,
      daysInMonth,
      status: 'ON_TRACK',
      label: 'Tanpa Target',
      badgeClass: 'bg-surface-elevated text-text-secondary border-border-subtle',
      barGradient: 'from-accent-magic to-sky-400',
      message: 'Tidak ada target koin minimal untuk periode ini.',
    }
  }

  if (earned >= safeTarget) {
    return {
      target: safeTarget,
      earned,
      targetPercent: 100,
      monthProgress,
      dayOfMonth,
      daysInMonth,
      status: 'ACHIEVED',
      label: '🎯 Target Tercapai',
      badgeClass: 'bg-status-success/15 text-status-success border-status-success/30',
      barGradient: 'from-emerald-500 to-teal-400',
      message: 'Luar biasa! Target minimal bulan ini telah tercapai.',
    }
  }

  if (dayOfMonth >= 20 && targetPercent < 50) {
    return {
      target: safeTarget,
      earned,
      targetPercent,
      monthProgress,
      dayOfMonth,
      daysInMonth,
      status: 'CRITICAL',
      label: '🚨 Di Bawah Target',
      badgeClass: 'bg-status-error/15 text-status-error border-status-error/30',
      barGradient: 'from-rose-500 to-red-500',
      message:
        'Peringatan: Capaian koinmu masih sangat rendah mendekati akhir bulan. Segera selesaikan misi harian agar tidak terkena evaluasi/pembekuan akun.',
    }
  }

  if (dayOfMonth >= 12 && targetPercent < monthProgress * 0.6) {
    return {
      target: safeTarget,
      earned,
      targetPercent,
      monthProgress,
      dayOfMonth,
      daysInMonth,
      status: 'BEHIND',
      label: '⚠️ Tertinggal',
      badgeClass: 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border-amber-500/30',
      barGradient: 'from-amber-500 to-orange-400',
      message:
        'Capaian koinmu tertinggal dibanding progres hari bulan ini. Kejar dengan menyelesaikan seluruh tugas harianmu.',
    }
  }

  return {
    target: safeTarget,
    earned,
    targetPercent,
    monthProgress,
    dayOfMonth,
    daysInMonth,
    status: 'ON_TRACK',
    label: '🟢 Sesuai Target',
    badgeClass: 'bg-status-success/10 text-status-success border-status-success/20',
    barGradient: 'from-emerald-500 to-teal-400',
    message: 'Performa bagus! Kamu berada di jalur yang tepat untuk memenuhi target bulan ini.',
  }
}
