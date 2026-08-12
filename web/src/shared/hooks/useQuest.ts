import { useState, useEffect, useCallback } from 'react'
import { questsApi } from '../lib/api'
import type { QuestWithChallenges, CompleteChallengeResult } from '../types'
import { useSession } from './useSession'

export function useQuest(questId?: number) {
  const { refreshProfile } = useSession()
  const [data, setData] = useState<QuestWithChallenges | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchQuest = useCallback(async () => {
    if (!questId || isNaN(questId)) {
      setLoading(false)
      return
    }
    setLoading(true)
    setError(null)
    try {
      const res = await questsApi.get(questId)
      setData(res)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load quest')
    } finally {
      setLoading(false)
    }
  }, [questId])

  useEffect(() => {
    fetchQuest()
  }, [fetchQuest])

  const startQuest = async () => {
    if (!questId) return
    setError(null)
    try {
      await questsApi.start(questId)
      await fetchQuest()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to start quest')
      throw e
    }
  }

  const completeChallenge = async (challengeId: number, payload?: { answer?: string, content?: string }): Promise<CompleteChallengeResult | null> => {
    if (!questId) return null
    setError(null)
    try {
      const result = await questsApi.completeChallenge(questId, challengeId, payload)
      if (result && result.quest) {
        setData(result.quest)
      } else {
        await fetchQuest()
      }
      // Quest completion grants +5 coins server-side; refresh session profile so
      // Home/Profile/Sidebar balance is not stale after CREATE_MEMORY / navigate away.
      if (result?.quest_completed) {
        await refreshProfile()
      }
      return result
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to complete challenge')
      throw e
    }
  }

  const selectBranch = async (branchSlug: string) => {
    if (!questId) return
    setError(null)
    try {
      const res = await questsApi.selectBranch(questId, branchSlug)
      await fetchQuest()
      return res
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to select branch')
      throw e
    }
  }

  return {
    quest: data,
    challenges: data?.challenges ?? [],
    loading,
    error,
    refresh: fetchQuest,
    startQuest,
    completeChallenge,
    selectBranch,
  }
}
