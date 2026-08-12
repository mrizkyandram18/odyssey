import { useState, useEffect, useCallback } from 'react'
import { MissionsApi } from '../lib/api'
import type { MissionWithChallenges, CompleteChallengeResult } from '../types'
import { useSession } from './useSession'

export function useMission(missionId?: number) {
  const { refreshProfile } = useSession()
  const [data, setData] = useState<MissionWithChallenges | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const fetchMission = useCallback(async () => {
    if (!missionId || isNaN(missionId)) {
      setLoading(false)
      return
    }
    setLoading(true)
    setError(null)
    try {
      const res = await MissionsApi.get(missionId)
      setData(res)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to load Mission')
    } finally {
      setLoading(false)
    }
  }, [missionId])

  useEffect(() => {
    fetchMission()
  }, [fetchMission])

  const startMission = async () => {
    if (!missionId) return
    setError(null)
    try {
      await MissionsApi.start(missionId)
      await fetchMission()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to start Mission')
      throw e
    }
  }

  const completeChallenge = async (exerciseId: number, payload?: { answer?: string, content?: string }): Promise<CompleteChallengeResult | null> => {
    if (!missionId) return null
    setError(null)
    try {
      const result = await MissionsApi.completeChallenge(missionId, exerciseId, payload)
      if (result && result.Mission) {
        setData(result.Mission)
      } else {
        await fetchMission()
      }
      // Mission completion grants +5 coins server-side; refresh session profile so
      // Home/Profile/Sidebar balance is not stale after CREATE_MEMORY / navigate away.
      if (result?.Mission_completed) {
        await refreshProfile()
      }
      return result
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to complete challenge')
      throw e
    }
  }

  const selectBranch = async (branchSlug: string) => {
    if (!missionId) return
    setError(null)
    try {
      const res = await MissionsApi.selectBranch(missionId, branchSlug)
      await fetchMission()
      return res
    } catch (e) {
      setError(e instanceof Error ? e.message : 'failed to select branch')
      throw e
    }
  }

  return {
    Mission: data,
    exercises: data?.exercises ?? [],
    loading,
    error,
    refresh: fetchMission,
    startMission,
    completeChallenge,
    selectBranch,
  }
}
