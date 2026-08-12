import { usePushSubscription } from '../../hooks/usePushSubscription'
import { Button } from '../atoms/Button'
import { Card } from '../atoms/Card'
import { Bell, BellOff, AlertTriangle } from 'lucide-react'

export function PushNotificationToggle() {
  const { status, error, subscribe, unsubscribe } = usePushSubscription()

  return (
    <Card className="p-6" data-testid="push-notification-settings">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="p-2 rounded-lg bg-accent-magic/10 text-accent-magic">
            <Bell size={24} />
          </div>
          <div>
            <h3 className="font-heading text-xl text-text-primary flex items-center gap-2">
              Push Notifications (PWA)
            </h3>
            <p className="text-xs text-text-secondary mt-1">
              Receive updates for daily turns &amp; Mission handoffs even when Odyssey is closed.
            </p>
          </div>
        </div>
      </div>

      <div className="mt-4 pt-4 border-t border-border-subtle flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="text-xs">
          {status === 'enabled' && (
            <span className="font-medium text-accent-nature flex items-center gap-1.5" data-testid="push-status-enabled">
              <span className="h-2 w-2 rounded-full bg-accent-nature animate-pulse" />
              Push notifications active
            </span>
          )}
          {status === 'permission_required' && (
            <span className="text-text-secondary" data-testid="push-status-required">
              Permission required to send Web Push notifications.
            </span>
          )}
          {status === 'blocked' && (
            <span className="text-accent-danger flex items-center gap-1" data-testid="push-status-blocked">
              <AlertTriangle size={14} />
              Notifications blocked in browser settings.
            </span>
          )}
          {status === 'unsupported' && (
            <span className="text-text-secondary italic" data-testid="push-status-unsupported">
              Web Push is not supported in this browser environment.
            </span>
          )}
          {status === 'loading' && (
            <span className="text-text-secondary animate-pulse" data-testid="push-status-loading">
              Checking notification status…
            </span>
          )}
        </div>

        <div>
          {status === 'enabled' && (
            <Button
              size="sm"
              variant="secondary"
              onClick={() => void unsubscribe()}
              className="flex items-center gap-2"
              data-testid="disable-push-btn"
            >
              <BellOff size={14} /> Disable Notifications
            </Button>
          )}
          {(status === 'permission_required' || status === 'error') && (
            <Button
              size="sm"
              variant="primary"
              onClick={() => void subscribe()}
              className="flex items-center gap-2"
              data-testid="enable-push-btn"
            >
              <Bell size={14} /> Enable Notifications
            </Button>
          )}
        </div>
      </div>

      {error && (
        <p className="mt-3 text-xs text-accent-danger" data-testid="push-error-msg">
          {error}
        </p>
      )}
    </Card>
  )
}
