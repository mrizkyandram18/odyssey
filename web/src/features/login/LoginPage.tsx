import { useState } from 'react'
import { Button } from '../../shared/components/atoms/Button'
import { Input } from '../../shared/components/atoms/Input'
import { useSession } from '../../shared/hooks/useSession'
import type { DevicePayload } from '../../shared/types'

export function LoginPage() {
  const { login } = useSession()
  const [uid, setUid] = useState('')
  const [credential, setCredential] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleLogin = async () => {
    setLoading(true)
    setError(null)
    try {
      const device: DevicePayload = {
        device_id: 'web-pwa',
        login_method: 'BOTH',
      }
      await login(uid, credential, device)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Login failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 p-6">
      <div className="w-full max-w-sm space-y-4">
        <h1 className="text-center text-xl font-bold">Odyssey</h1>
        <p className="text-center text-sm text-muted-foreground">
          Sign in with Gatekeeper BOTH login
        </p>
        <div className="space-y-3">
          <Input
            label="UID"
            placeholder="Family member ID"
            value={uid}
            onChange={(e) => setUid(e.target.value)}
          />
          <Input
            label="Password (PIN)"
            type="password"
            placeholder="••••••"
            value={credential}
            onChange={(e) => setCredential(e.target.value)}
          />
        </div>
        {error && <p className="text-xs text-error">{error}</p>}
        <Button
          className="w-full"
          isLoading={loading}
          onClick={handleLogin}
        >
          Sign In
        </Button>
      </div>
    </div>
  )
}
