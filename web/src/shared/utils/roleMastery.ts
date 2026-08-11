import type { Role } from '../types'

export interface RoleMastery {
  title: string
  flavor: string
}

export function getRoleMastery(role: Role | string | undefined | null, level: number): RoleMastery {
  const normRole = (role || 'SEEKER').toUpperCase()
  const safeLevel = Math.max(1, level || 1)

  if (normRole === 'SEEKER') {
    if (safeLevel >= 10) {
      return { title: 'Master Seeker', flavor: 'Master Seeker whose gaze pierces through illusions and uncovers the deepest secrets.' }
    }
    if (safeLevel >= 5) {
      return { title: 'Adept Seeker', flavor: 'Keen-eyed tracker who uncovers the realm\'s hidden mysteries and subtle details.' }
    }
    return { title: 'Novice Seeker', flavor: 'Curious explorer who seeks out hidden details, riddles, and lore.' }
  }

  if (normRole === 'BUILDER') {
    if (safeLevel >= 10) {
      return { title: 'Master Builder', flavor: 'Master Builder who shapes the very fabric of the adventure and leaves a lasting legacy.' }
    }
    if (safeLevel >= 5) {
      return { title: 'Adept Builder', flavor: 'Skilled artisan who pieces together complex puzzles and crafts clever solutions.' }
    }
    return { title: 'Novice Builder', flavor: 'Creative craftsman who solves puzzles and constructs family artifacts.' }
  }

  if (normRole === 'GUIDE') {
    if (safeLevel >= 10) {
      return { title: 'Master Guide', flavor: 'Master Guide whose wisdom leads the crew through any storm and inspires greatness.' }
    }
    if (safeLevel >= 5) {
      return { title: 'Adept Guide', flavor: 'Seasoned pathfinder who navigates treacherous routes and coordinates complex quests.' }
    }
    return { title: 'Novice Guide', flavor: 'Helpful companion who points the way and mentors the crew.' }
  }

  // Fallback for unknown role
  return { title: `Level ${safeLevel} Explorer`, flavor: 'An intrepid explorer ready for any adventure.' }
}
