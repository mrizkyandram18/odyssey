import { useState, useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import confetti from 'canvas-confetti'
import { chestsApi } from '../../shared/lib/api'
import type { OpenResult } from '../../shared/types'
import { CollectionCard } from '../collections/components/CollectionCard'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'

export function GiftOpeningPage() {
  const { chestId } = useParams<{ chestId: string }>()
  const [result, setResult] = useState<OpenResult | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const open = async () => {
      if (!chestId) return
      setLoading(true)
      setError(null)
      try {
        const data = await chestsApi.open(Number(chestId))
        setResult(data)
        
        // Trigger confetti when successful
        confetti({
          particleCount: 150,
          spread: 80,
          origin: { y: 0.5 },
          colors: ['#06b6de', '#f59e0b', '#10b981']
        })
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed to open chest')
      } finally {
        setLoading(false)
      }
    }
    open()
  }, [chestId])

  if (loading) {
    return (
      <div className="flex flex-col max-w-2xl mx-auto items-center justify-center gap-4 p-8 min-h-[60vh]">
        <motion.div 
          animate={{ scale: [1, 1.1, 1], rotate: [0, 5, -5, 0] }}
          transition={{ repeat: Infinity, duration: 1.5 }}
          className="text-6xl drop-shadow-[0_0_20px_rgba(245,158,11,0.5)]"
        >
          📦
        </motion.div>
        <p className="text-lg text-text-secondary font-medium">Opening chest...</p>
      </div>
    )
  }

  if (error || !result) {
    return (
      <div className="flex flex-col gap-6 max-w-2xl mx-auto p-4">
        <Link to="/gifts" className="text-sm text-text-secondary hover:text-text-primary">← Back to Gifts</Link>
        <Card className="bg-accent-danger/10 border-accent-danger/30 text-center">
          <p className="text-lg font-medium text-accent-danger">{error || 'Failed to open chest'}</p>
        </Card>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6 max-w-2xl mx-auto p-4 pb-safe">
      <Link to="/" className="text-sm text-text-secondary hover:text-text-primary">← Back to Adventure Hub</Link>

      <Card className="flex flex-col items-center gap-4 p-8 text-center border-accent-reward bg-accent-reward/5 shadow-[0_0_30px_rgba(245,158,11,0.15)] relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-t from-accent-reward/20 to-transparent opacity-50"></div>
        <motion.div 
          initial={{ scale: 0, rotate: -180 }}
          animate={{ scale: 1, rotate: 0 }}
          transition={{ type: "spring", bounce: 0.5 }}
          className="text-7xl drop-shadow-[0_0_30px_rgba(245,158,11,0.8)] relative z-10"
        >
          {result.chest.icon}
        </motion.div>
        <div className="relative z-10">
          <h1 className="font-heading text-3xl text-text-primary mb-2">{result.chest.name} Opened!</h1>
          <p className="text-text-secondary">{result.chest.description}</p>
        </div>
      </Card>

      <div className="flex flex-col gap-4">
        <h2 className="font-heading text-xl text-text-primary border-b border-border-subtle pb-2">Rewards Discovered</h2>
        <motion.div 
          initial="hidden"
          animate="show"
          variants={{
            hidden: { opacity: 0 },
            show: { opacity: 1, transition: { staggerChildren: 0.15 } }
          }}
          className="flex flex-col gap-3"
        >
          {result.rewards.map((reward, idx) => (
            <motion.div key={idx} variants={{ hidden: { opacity: 0, x: -20 }, show: { opacity: 1, x: 0 } }}>
              <CollectionCard
                relic={{
                  collection_id: 0,
                  collection_slug: reward.collection_slug,
                  name: reward.name,
                  description: '',
                  journey: '',
                  rarity: reward.rarity,
                  image: '',
                  concept: '',
                  owned_count: 1,
                  is_new: reward.is_new,
                  discovered_at: new Date().toISOString(),
                  created_at: new Date().toISOString(),
                }}
              />
            </motion.div>
          ))}
        </motion.div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <Card className="flex flex-col items-center justify-center p-4">
          <span className="text-sm text-text-secondary mb-1">New Discoveries</span>
          <span className="text-2xl font-bold text-accent-magic">{result.new_count}</span>
        </Card>
        <Card className="flex flex-col items-center justify-center p-4">
          <span className="text-sm text-text-secondary mb-1">Duplicates</span>
          <span className="text-2xl font-bold text-text-primary">{result.duplicate_count}</span>
        </Card>
      </div>

      <Link to="/collections" className="mt-4">
        <Button size="lg" className="w-full">
          View Collection
        </Button>
      </Link>
    </div>
  )
}
