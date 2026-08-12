import { useState, useEffect } from 'react';
import { apiClient } from '../../shared/lib/api';
import { Button } from '../../shared/components/atoms/Button';

interface ActivityView {
  id: number;
  title: string;
  type: string;
  question: string;
  options: string[];
  completed: boolean;
  xp_reward: number;
}

interface ActivityResult {
  correct: boolean;
  completed: boolean;
  explanation: string;
  xp_awarded?: number;
}

export const DailyActivitySection = () => {
  const [activity, setActivity] = useState<ActivityView | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [selectedOption, setSelectedOption] = useState<string | null>(null);
  const [result, setResult] = useState<ActivityResult | null>(null);

  const fetchActivity = async () => {
    try {
      const data = await apiClient.get<ActivityView>('/api/daily-activities/today');
      setActivity(data);
    } catch (e: any) {
      if (e.response?.status !== 404) {
        console.error('Failed to fetch daily activity', e);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivity();
  }, []);

  const handleSubmit = async () => {
    if (!selectedOption || !activity) return;
    setSubmitting(true);
    setResult(null);
    try {
      const data = await apiClient.post<ActivityResult>(`/api/daily-activities/${activity.id}/complete`, {
        answer: selectedOption,
      });
      setResult(data);
      if (data.completed) {
        setActivity((prev) => prev ? { ...prev, completed: true } : null);
      }
    } catch (e) {
      console.error('Failed to submit activity', e);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) return <div className="p-4 rounded-xl border border-border/50 animate-pulse bg-surface-base h-32" />;
  if (!activity) return null;

  if (activity.completed) {
    return (
      <div className="p-4 rounded-xl border border-success/30 bg-success/10 flex flex-col items-center text-center gap-2">
        <h3 className="font-semibold text-text-primary text-lg">✅ Aktivitas Hari Ini Selesai</h3>
        <p className="text-sm text-success font-medium">+{activity.xp_reward} Koin 🪙</p>
        <p className="text-sm text-text-secondary mt-1">Kembali lagi besok untuk aktivitas berikutnya.</p>
      </div>
    );
  }

  return (
    <div className="p-5 rounded-xl border border-primary/20 bg-surface-elevated flex flex-col gap-4">
      <div>
        <h3 className="font-bold text-text-primary text-lg mb-1">Aktivitas Hari Ini</h3>
        <div className="flex items-center gap-2 text-xs text-text-secondary">
          <span className="px-2 py-0.5 rounded-full bg-primary/10 text-primary font-medium">{activity.title}</span>
          <span>Sekitar 30 detik</span>
        </div>
      </div>
      
      <p className="text-text-primary font-medium">{activity.question}</p>
      
      <div className="flex flex-col gap-2 mt-2">
        {activity.options.map((opt) => (
          <button
            key={opt}
            onClick={() => setSelectedOption(opt)}
            disabled={submitting}
            className={`w-full text-left p-3 rounded-lg border transition-colors ${
              selectedOption === opt 
                ? 'border-primary bg-primary/10 text-primary font-medium' 
                : 'border-border hover:border-primary/50 text-text-secondary'
            }`}
          >
            {opt}
          </button>
        ))}
      </div>

      {result && (
        <div className={`p-3 rounded-lg mt-2 ${result.correct ? 'bg-success/20 text-success' : 'bg-error/20 text-error'}`}>
          <p className="font-semibold">{result.correct ? '✅ Benar!' : '❌ Belum tepat.'}</p>
          <p className="text-sm mt-1 text-text-primary">{result.explanation}</p>
          {result.correct && result.xp_awarded && (
            <p className="text-sm font-bold mt-1">+{result.xp_awarded} Koin 🪙</p>
          )}
          {!result.correct && (
            <p className="text-sm mt-1 italic opacity-80">Coba lagi.</p>
          )}
        </div>
      )}

      <Button 
        onClick={handleSubmit} 
        disabled={!selectedOption || submitting}
        className="w-full mt-2"
        variant={selectedOption ? 'primary' : 'secondary'}
      >
        Jawab
      </Button>
    </div>
  );
};
