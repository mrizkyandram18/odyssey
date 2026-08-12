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
                <div className="flex flex-col h-full justify-between">
                  <div className="text-center mb-6">
                    <span className="text-5xl block mb-4">✨</span>
                    <h2 className="font-heading text-3xl text-text-primary mb-4">Selamat datang di Odyssey!</h2>
                    <p className="text-sm text-text-secondary leading-relaxed">
                      Odyssey membantu kamu belajar hal-hal penting untuk kehidupan sehari-hari, sambil melakukannya bersama keluarga.
                    </p>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-lg" onClick={handleNext}>
                    Lanjut
                  </Button>
                </div>
              )}

              {step === 2 && (
                <div className="flex flex-col h-full justify-between">
                  <div className="mb-6">
                    <h2 className="font-heading text-2xl text-text-primary mb-6 text-center">Kamu akan belajar tentang:</h2>
                    <div className="flex flex-col gap-4">
                      <div>
                        <h3 className="font-bold text-sm text-text-primary">🔐 Keamanan digital</h3>
                        <p className="text-xs text-text-secondary">Cara mengenali penipuan dan menjaga akunmu.</p>
                      </div>
                      <div>
                        <h3 className="font-bold text-sm text-text-primary">💰 Keuangan</h3>
                        <p className="text-xs text-text-secondary">Cara mengatur uang dan membuat keputusan keuangan.</p>
                      </div>
                      <div>
                        <h3 className="font-bold text-sm text-text-primary">💼 Dunia kerja</h3>
                        <p className="text-xs text-text-secondary">Cara membuat CV dan menghadapi wawancara.</p>
                      </div>
                      <div>
                        <h3 className="font-bold text-sm text-text-primary">📱 Produktivitas</h3>
                        <p className="text-xs text-text-secondary">Cara mengatur waktu dan menyelesaikan prioritas.</p>
                      </div>
                    </div>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-lg" onClick={handleNext}>
                    Lanjut
                  </Button>
                </div>
              )}

              {step === 3 && (
                <div className="flex flex-col h-full justify-between">
                  <div className="mb-6">
                    <h2 className="font-heading text-2xl text-text-primary mb-4 text-center">Cara belajarnya sederhana:</h2>
                    <ol className="list-decimal list-inside text-sm text-text-secondary flex flex-col gap-3 ml-2 mb-6">
                      <li>Pelajari materi singkat (5–10 menit)</li>
                      <li>Kerjakan latihan</li>
                      <li>Lihat hasil dan penjelasannya</li>
                      <li>Dapatkan poin dan hadiah</li>
                    </ol>
                    <p className="text-sm text-text-secondary text-center italic">
                      Keluargamu juga bisa melihat perkembanganmu.
                    </p>
                  </div>
                  <Button size="lg" className="w-full text-lg shadow-lg" onClick={handleStart}>
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
