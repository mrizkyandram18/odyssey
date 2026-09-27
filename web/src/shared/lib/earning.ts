export function isEarningCapError(err: unknown): boolean {
  const msg = (err instanceof Error ? err.message : String(err ?? '')).toLowerCase()
  return msg.includes('earning_cap_reached') || msg.includes('p0016') || msg.includes('batas earning')
}

export const EARNING_CAP_MESSAGE =
  'Batas earning bulan berjalan tercapai. Kamu tidak dapat memperoleh Koin lagi sampai bulan berikutnya. Saldo tetap aman.'
