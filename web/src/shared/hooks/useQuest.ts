export function useQuest(questId?: number) {
  void questId
  return {
    quest: null,
    challenges: [],
    loading: false,
    error: null,
  }
}
