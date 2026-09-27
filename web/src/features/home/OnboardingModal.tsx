import { useState, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Card } from '../../shared/components/atoms/Card'
import { Button } from '../../shared/components/atoms/Button'

export function OnboardingModal() {
  const [isOpen, setIsOpen] = useState(false)
  const [step, setStep] = useState(1)

  useEffect(() => {
    const hasOnboarded = localStorage.getItem('odyssey_onboarded')
    if (!hasOnboarded) {
      setIsOpen(true)
    }
  }, [])

  const handleNext = () => setStep(step + 1)
  const handleStart = () => {
    localStorage.setItem('odyssey_onboarded', 'true')
    setIsOpen(false)
  }

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
          <motion.div
            initial={{ opacity: 0, scale: 0.9, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 20 }}
            className="w-full max-w-md"
          >
            <Card className="p-6 md:p-8 shadow-xl overflow-hidden relative">
              
              {step === 1 && (
                <div className="flex flex-col h-full justify-between items-center text-center">
                  <div className="mb-6 flex flex-col items-center">
                    <span className="text-6xl block mb-6">📝</span>
                    <h2 className="font-heading text-2xl font-bold text-text-primary mb-4 leading-snug">
                      1. Kerjakan tugas harian
                    </h2>
                    <p className="text-sm text-text-secondary leading-relaxed max-w-xs">
                      Setiap hari ada daftar tugas berurutan: video, kuis, foto, atau jawaban singkat. Selesaikan dari langkah pertama.
                    </p>
                    <p className="text-xs text-text-secondary leading-relaxed max-w-xs mt-3">
                      Sebagian tugas dinilai otomatis, sebagian menunggu verifikasi admin sebelum reward masuk.
                    </p>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-sm" onClick={handleNext}>
                    Selanjutnya
                  </Button>
                </div>
              )}

              {step === 2 && (
                <div className="flex flex-col h-full justify-between text-center">
                  <div className="mb-6">
                    <span className="text-6xl block mb-6">🪙</span>
                    <h2 className="font-heading text-2xl font-bold text-text-primary mb-4 leading-snug">
                      2. Dapatkan Koin & Bintang
                    </h2>
                    <p className="text-sm text-text-secondary leading-relaxed max-w-xs mx-auto">
                      Tugas yang disetujui memberi <strong className="text-text-primary">Koin</strong> untuk ditukar dan <strong className="text-text-primary">Bintang (XP)</strong> untuk naik tingkat.
                    </p>
                    <p className="text-xs text-text-secondary leading-relaxed max-w-xs mx-auto mt-3">
                      Besaran Koin mengikuti target dan batas periode yang berlaku, jadi angka per tugas bisa berbeda dari waktu ke waktu.
                    </p>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-sm" onClick={handleNext}>
                    Selanjutnya
                  </Button>
                </div>
              )}

              {step === 3 && (
                <div className="flex flex-col h-full justify-between text-center">
                  <div className="mb-6">
                    <span className="text-6xl block mb-6">💸</span>
                    <h2 className="font-heading text-2xl font-bold text-text-primary mb-2">3. Tukarkan Koin menjadi uang.</h2>
                    <div className="mt-6 flex items-center justify-center gap-2 text-xs md:text-sm font-semibold text-text-secondary flex-wrap">
                       <span className="bg-surface border border-border-subtle px-3 py-1 rounded-full">Selesaikan Tugas</span>
                       <span>→</span>
                       <span className="bg-amber-500/10 text-amber-700 dark:text-amber-300 px-3 py-1 rounded-full font-bold">Koin Masuk</span>
                       <span>→</span>
                       <span className="bg-accent-magic/10 text-accent-magic px-3 py-1 rounded-full font-bold">Cairkan di Penukaran Koin</span>
                    </div>
                    <p className="text-sm text-text-secondary mt-6">
                      Buka halaman <strong className="text-text-primary">Penukaran Koin</strong> untuk melihat saldo, minimum pencairan, dan status pengajuanmu.
                    </p>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-sm" onClick={handleStart}>
                    Mulai Sekarang
                  </Button>
                </div>
              )}
            </Card>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  )
}
