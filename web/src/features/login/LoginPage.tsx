import { useState } from 'react'
import { Button } from '../../shared/components/atoms/Button'
import { Input } from '../../shared/components/atoms/Input'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import type { DevicePayload } from '../../shared/types'

export function LoginPage() {
  const { login } = useSession()
  const [username, setUsername] = useState('')
  const [credential, setCredential] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError(null)
    try {
      const device: DevicePayload = {
        device_id: 'web-pwa',
        login_method: 'BOTH',
      }
      await login(username, credential, device)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login gagal')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-bg-app">
      {/* Atmospheric Background */}
      <div className="absolute inset-0 z-0">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(6,182,222,0.15)_0%,rgba(10,12,21,1)_100%)]"></div>
        <div className="absolute top-0 left-0 right-0 h-96 bg-gradient-to-b from-accent-nature/10 to-transparent"></div>
        <div className="absolute bottom-0 left-0 right-0 h-64 bg-gradient-to-t from-bg-realm to-transparent"></div>
        {/* Subtle decorative stars/particles could go here */}
      </div>

      <div className="relative z-10 w-full max-w-md p-6">
        <div className="mb-10 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-surface-elevated border border-accent-magic/30 shadow-[0_0_20px_rgba(6,182,222,0.2)]">
            <span className="text-3xl">🧭</span>
          </div>
          <h1 className="font-heading text-4xl text-text-primary mb-2 tracking-wide">Odyssey</h1>
          <p className="text-base text-text-secondary">Krumu sedang menunggu petualangan berikutnya.</p>
        </div>

        <Card className="shadow-2xl shadow-black/50 border-border-subtle/50 bg-surface-glass">
          <form onSubmit={handleLogin} className="flex flex-col gap-6">
            <div className="space-y-4">
              <Input
                label="Nama Penjelajah"
                placeholder="Masukkan nama pengguna"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
                required
              />
              <Input
                label="Kata Rahasia"
                type="password"
                placeholder="••••••••"
                value={credential}
                onChange={(e) => setCredential(e.target.value)}
                autoComplete="current-password"
                required
              />
            </div>
            
            {error && (
              <div className="rounded-md bg-accent-danger/10 border border-accent-danger/20 p-3">
                <p className="text-sm font-medium text-accent-danger text-center">{error}</p>
              </div>
            )}
            
            <Button
              type="submit"
              size="lg"
              className="w-full mt-2"
              isLoading={loading}
            >
              Mulai Petualangan
            </Button>
          </form>
        </Card>
      </div>
    </div>
  )
}
