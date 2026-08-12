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
            <Card className="p-6 md:p-8 bg-surface-elevated border-accent-magic/30 shadow-2xl overflow-hidden relative">
              <div className="absolute top-0 right-0 w-32 h-32 bg-accent-magic/10 rounded-bl-full -z-10" />
              
              {step === 1 && (
                <div className="flex flex-col h-full justify-between items-center text-center">
                  <div className="mb-6 flex flex-col items-center">
                    <span className="text-6xl block mb-6">👨‍👩‍👧‍👦</span>
                    <h2 className="font-heading text-2xl font-bold text-text-primary mb-4 leading-snug">
                      Odyssey membantu kamu belajar hal yang berguna untuk kehidupan sehari-hari.
                    </h2>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-sm" onClick={handleNext}>
                    Selanjutnya
                  </Button>
                </div>
              )}

              {step === 2 && (
                <div className="flex flex-col h-full justify-between text-center">
                  <div className="mb-6">
                    <h2 className="font-heading text-xl font-bold text-text-primary mb-6">Topik yang akan dipelajari:</h2>
                    <div className="flex flex-col gap-5 text-left max-w-xs mx-auto">
                      <div className="flex items-center gap-4 bg-surface p-3 rounded-2xl border border-border-subtle shadow-sm">
                        <span className="text-3xl bg-orange-100 p-2 rounded-xl">🛡️</span>
                        <div>
                          <h3 className="font-bold text-text-primary">Aman di internet</h3>
                          <p className="text-xs text-text-secondary mt-1">Belajar bijak dan aman online.</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-4 bg-surface p-3 rounded-2xl border border-border-subtle shadow-sm">
                        <span className="text-3xl bg-green-100 p-2 rounded-xl">💰</span>
                        <div>
                          <h3 className="font-bold text-text-primary">Mengatur uang</h3>
                          <p className="text-xs text-text-secondary mt-1">Kelola keuangan untuk masa depan.</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-4 bg-surface p-3 rounded-2xl border border-border-subtle shadow-sm">
                        <span className="text-3xl bg-blue-100 p-2 rounded-xl">💼</span>
                        <div>
                          <h3 className="font-bold text-text-primary">Siap kerja</h3>
                          <p className="text-xs text-text-secondary mt-1">Persiapkan karir dan keterampilan.</p>
                        </div>
                      </div>
                    </div>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-sm" onClick={handleNext}>
                    Selanjutnya
                  </Button>
                </div>
              )}

              {step === 3 && (
                <div className="flex flex-col h-full justify-between text-center">
                  <div className="mb-6">
                    <span className="text-6xl block mb-6">⏱️</span>
                    <h2 className="font-heading text-2xl font-bold text-text-primary mb-2">Setiap hari cukup beberapa menit.</h2>
                    <div className="mt-6 flex items-center justify-center gap-2 text-sm font-semibold text-text-secondary">
                       <span className="bg-surface border border-border-subtle px-3 py-1 rounded-full">Belajar</span>
                       <span>→</span>
                       <span className="bg-surface border border-border-subtle px-3 py-1 rounded-full">Latihan</span>
                       <span>→</span>
                       <span className="bg-accent-reward/10 text-accent-reward px-3 py-1 rounded-full">Poin</span>
                    </div>
                    <p className="text-sm text-text-secondary mt-6">
                      Lihat perkembanganmu bersama keluarga!
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
