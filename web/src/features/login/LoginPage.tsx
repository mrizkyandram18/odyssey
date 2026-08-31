import { useState } from 'react'
import { Button } from '../../shared/components/atoms/Button'
import { Input } from '../../shared/components/atoms/Input'
import { Card } from '../../shared/components/atoms/Card'
import { useSession } from '../../shared/hooks/useSession'
import type { DevicePayload } from '../../shared/types'

import { getOrCreateDeviceId } from '../../shared/lib/device'

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
        device_id: getOrCreateDeviceId(),
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
    <div className="flex min-h-screen items-center justify-center bg-bg-app px-4 py-8">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-surface border border-border-subtle shadow-sm">
            <span className="text-2xl">🧭</span>
          </div>
          <h1 className="text-2xl font-bold text-text-primary tracking-tight">Odyssey</h1>
          <p className="text-sm text-text-secondary mt-1">Masuk untuk melanjutkan aktivitas belajar</p>
        </div>

        <Card className="shadow-sm">
          <form onSubmit={handleLogin} className="flex flex-col gap-5">
            <div className="space-y-4">
              <Input
                label="Nama Pengguna"
                placeholder="Masukkan nama pengguna"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                autoComplete="username"
                required
              />
              <Input
                label="Kata Sandi"
                type="password"
                placeholder="••••••••"
                value={credential}
                onChange={(e) => setCredential(e.target.value)}
                autoComplete="current-password"
                required
              />
            </div>
            
            {error && (
              <div className="rounded-xl bg-accent-danger/10 border border-accent-danger/20 p-3">
                <p className="text-xs font-semibold text-accent-danger text-center">{error}</p>
              </div>
            )}
            
            <Button
              type="submit"
              size="lg"
              className="w-full"
              isLoading={loading}
            >
              Masuk
            </Button>
            <p className="text-center text-[11px] text-text-secondary">
              Login menggunakan akun terdaftar
            </p>
          </form>
        </Card>
      </div>
    </div>
  )
}
