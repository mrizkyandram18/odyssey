#!/usr/bin/env node
/**
 * Odyssey Phase 3 — read-only retention report (operator CLI, no app changes).
 *
 * Measures D+1 / D+3 / D+7 RETURN as user-initiated re-engagement, computed
 * only from existing tables. No writes, no migrations, no new endpoints.
 *
 * Metric definitions (source-backed, see header notes):
 * - Cohort APPROVED (primary "completed"): users with >=1 submission whose
 *   status=APPROVED and reviewed_at falls on date D (system timezone).
 *   Rationale: rewarded completion is the product's own "selesai" semantics
 *   (UNIQUE(task_id,user_uid) + P0004 exactly-once approval).
 * - Cohort ATTEMPTED (secondary): users with >=1 submission created on D.
 * - Cohort PENDING (point-in-time): submissions created on D that are still
 *   PENDING now (reviewed_at IS NULL). Status is mutable, so this cohort
 *   describes the present, not a historical state.
 * - Cohort CLAIM: users with >=1 claim created on D.
 * - RETURN on D+n: >=1 user-initiated event on that date — submission
 *   created OR claim created. Approval/review is admin-driven and NEVER
 *   counts as user return. Streak snapshot cannot do history (overwritten
 *   in place, no history table) and is excluded.
 *
 * Day boundaries: all TIMESTAMPTZ values are absolute instants; they are
 * bucketed into calendar dates in --tz (default Asia/Jakarta, matching the
 * product's default system timezone). Known limitation: the codebase has no
 * single canonical day boundary (DB CURRENT_DATE for streak, configurable tz
 * for target periods/tickets, Go loc for today-list) — see final report.
 *
 * Privacy: selects only (user_uid, status, created_at, reviewed_at). Never
 * selects payload, admin_notes, target_value, names, or balances. Same access
 * level as existing operator verify scripts (service-role key, server-side).
 * Output shows counts + truncated UIDs for auditability, descriptive only.
 *
 * Usage:
 *   SUPABASE_URL=... SUPABASE_SERVICE_KEY=... node scripts/retention-report.cjs \
 *     --date=2026-09-10 [--tz=Asia/Jakarta] [--family-id=...]
 *   If --date is omitted, defaults to 7 days ago (so the D+7 window is complete).
 */

'use strict'

function parseArgs (argv) {
  const out = {}
  for (const a of argv) {
    const m = /^--([^=]+)=(.*)$/.exec(a)
    if (m) out[m[1]] = m[2]
  }
  return out
}

function dayInTz (iso, tz) {
  // en-CA yields YYYY-MM-DD; hour12 irrelevant when only Y/M/D requested.
  return new Intl.DateTimeFormat('en-CA', { timeZone: tz, year: 'numeric', month: '2-digit', day: '2-digit' }).format(new Date(iso))
}

function todayInTz (tz) {
  return dayInTz(new Date().toISOString(), tz)
}

function addDays (yyyyMmDd, n) {
  const [y, m, d] = yyyyMmDd.split('-').map(Number)
  const dt = new Date(Date.UTC(y, m - 1, d))
  dt.setUTCDate(dt.getUTCDate() + n)
  return dt.toISOString().slice(0, 10)
}

async function getAll (base, table, params) {
  // Paginated PostgREST GET. Read-only: never POST/PATCH/DELETE.
  const rows = []
  const limit = 1000
  let offset = 0
  for (;;) {
    const qs = new URLSearchParams({ ...params, order: 'id.asc', limit: String(limit), offset: String(offset) })
    const res = await fetch(`${base}/rest/v1/${table}?${qs.toString()}`, {
      method: 'GET',
      headers: {
        apikey: process.env.SUPABASE_SERVICE_KEY,
        Authorization: `Bearer ${process.env.SUPABASE_SERVICE_KEY}`
      }
    })
    if (!res.ok) throw new Error(`GET ${table} failed: ${res.status} ${await res.text()}`)
    const page = await res.json()
    rows.push(...page)
    if (page.length < limit) break
    offset += limit
  }
  return rows
}

function shortUid (uid) {
  return String(uid).slice(0, 8)
}

function reportCohort (label, cohortSet, activityByUser, D, days, tz) {
  const lines = []
  lines.push(`\n## ${label}: n=${cohortSet.size}`)
  if (cohortSet.size === 0) {
    lines.push('NOT AVAILABLE — cohort kosong pada tanggal ini (tidak dapat dihitung).')
    return lines
  }
  for (const n of days) {
    const date = addDays(D, n)
    const returned = [...cohortSet].filter((u) => activityByUser.get(u)?.has(date))
    const pct = ((returned.length / cohortSet.size) * 100).toFixed(1)
    lines.push(`D+${n} (${date}): ${returned.length} / ${cohortSet.size} users (${pct}%) [${returned.map(shortUid).join(', ')}]`)
  }
  lines.push(`(Descriptive only — observed association, NOT causal. Timezone: ${tz}.)`)
  return lines
}

async function main () {
  const args = parseArgs(process.argv.slice(2))
  const tz = args.tz || 'Asia/Jakarta'
  // Validate tz early so bucketing never silently falls back.
  try { new Intl.DateTimeFormat('en-CA', { timeZone: tz }).format(new Date()) } catch { throw new Error(`invalid --tz: ${tz}`) }
  if (!process.env.SUPABASE_URL || !process.env.SUPABASE_SERVICE_KEY) {
    throw new Error('Missing SUPABASE_URL / SUPABASE_SERVICE_KEY env (same as existing verify scripts).')
  }
  const D = args.date || addDays(todayInTz(tz), -7)
  if (!/^\d{4}-\d{2}-\d{2}$/.test(D)) throw new Error(`invalid --date (expected YYYY-MM-DD): ${D}`)
  const base = process.env.SUPABASE_URL.replace(/\/$/, '')
  const days = [1, 3, 7]
  const lastDay = addDays(D, 7)

  // Minimal columns only (privacy: no payload/notes/targets/names/balances).
  const [subs, claims, profiles] = await Promise.all([
    getAll(base, 'odyssey_task_submissions', { select: 'user_uid,status,created_at,reviewed_at' }),
    getAll(base, 'odyssey_claims', { select: 'user_uid,status,created_at' }),
    args['family-id']
      ? getAll(base, 'odyssey_user_profiles', { select: 'uid,created_at,family_id', family_id: `eq.${args['family-id']}` })
      : getAll(base, 'odyssey_user_profiles', { select: 'uid,created_at' })
  ])
  const familyUids = args['family-id'] ? new Set(profiles.map((p) => p.uid)) : null
  const inScope = (u) => !familyUids || familyUids.has(u)

  // Users existing at cohort time (denominator hygiene for future extensions;
  // cohorts below are event-defined so this is informational).
  const existingByD7 = new Set(profiles.filter((p) => dayInTz(p.created_at, tz) <= lastDay).map((p) => p.uid))

  const approved = new Set()
  const attempted = new Set()
  const pending = new Set()
  const claimed = new Set()
  // User-initiated activity by user -> set of dates (the ONLY return signal).
  const activity = new Map()
  const mark = (u, date) => {
    if (!activity.has(u)) activity.set(u, new Set())
    activity.get(u).add(date)
  }

  for (const s of subs) {
    if (!inScope(s.user_uid)) continue
    const createdDay = dayInTz(s.created_at, tz)
    if (createdDay === D) {
      attempted.add(s.user_uid)
      if (s.status === 'PENDING' && !s.reviewed_at) pending.add(s.user_uid)
    }
    if (s.status === 'APPROVED' && s.reviewed_at && dayInTz(s.reviewed_at, tz) === D) {
      approved.add(s.user_uid)
    }
    mark(s.user_uid, createdDay) // submission created = user-initiated
  }
  for (const c of claims) {
    if (!inScope(c.user_uid)) continue
    const createdDay = dayInTz(c.created_at, tz)
    if (createdDay === D) claimed.add(c.user_uid)
    mark(c.user_uid, createdDay) // claim created = user-initiated
  }

  const out = []
  out.push('# Odyssey retention report (read-only, existing data only)')
  out.push(`Cohort date D: ${D} (tz ${tz})${args['family-id'] ? ` | family: ${args['family-id']}` : ''}`)
  out.push(`Observation window: ${D} → ${lastDay}`)
  out.push(`Users existing by D+7: ${existingByD7.size} (informational; cohorts are event-defined)`)
  out.push('Return = submission-created OR claim-created on D+n (user-initiated). Approvals never count as return.')
  for (const [label, set] of [
    ['Cohort APPROVED (reviewed_at on D)', approved],
    ['Cohort ATTEMPTED (submitted on D)', attempted],
    ['Cohort PENDING (created D, still pending now)', pending],
    ['Cohort CLAIM (claimed on D)', claimed]
  ]) {
    out.push(...reportCohort(label, set, activity, D, days, tz))
  }
  out.push('\nLimitations: visit-level return (app-open) NOT AVAILABLE — no login/view persistence; logs are stdout-only and ephemeral; metrics reset on boot. PENDING cohort is point-in-time (status mutable). Day bucketing uses --tz; DB streak uses CURRENT_DATE (session TZ) — boundaries may differ by hours. Small family scale: show n/N, no causal claims.')
  console.log(out.join('\n'))
}

if (require.main === module) {
  main().catch((e) => { console.error(`retention-report: ${e.message}`); process.exit(1) })
}

module.exports = { dayInTz, addDays, reportCohort, main }
