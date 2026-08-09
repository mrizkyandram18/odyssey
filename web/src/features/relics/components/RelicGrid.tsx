import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { InventoryItem } from '../../../shared/types'
import { RelicCard } from './RelicCard'

export interface RelicGridProps {
  relics: InventoryItem[]
  onRelicClick?: (relic: InventoryItem) => void
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.05 }
  }
}

const itemVariant = {
  hidden: { opacity: 0, y: 20 },
  show: { opacity: 1, y: 0 }
}

export function RelicGrid({ relics, onRelicClick }: RelicGridProps) {
  if (relics.length === 0) {
    return null
  }

  return (
    <motion.div 
      variants={container}
      initial="hidden"
      animate="show"
      className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4 md:gap-6"
    >
      {relics.map((relic) => (
        <motion.div key={relic.relic_slug} variants={itemVariant}>
          <Link to={`/relics/${relic.relic_slug}`} className="block">
            <RelicCard
              relic={relic}
              onClick={() => onRelicClick?.(relic)}
            />
          </Link>
        </motion.div>
      ))}
    </motion.div>
  )
}
