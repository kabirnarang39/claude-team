import React from 'react'
import { LimitExceededBody, QuotaExceededBody, RateLimitBody } from '../../api/client'

interface LimitToastProps {
  error: LimitExceededBody | QuotaExceededBody | RateLimitBody | null
  onClose: () => void
  onUpgrade?: () => void
}

function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(new Date(dateStr))
}

function getMessage(error: LimitExceededBody | QuotaExceededBody | RateLimitBody): string {
  switch (error.error) {
    case 'user_limit_exceeded':
      return `User limit reached (${error.current}/${error.limit}). Upgrade to Pro for up to 50 users.`
    case 'api_quota_exceeded':
      return `API quota exceeded. Resets on ${formatDate(error.reset_at)}.`
    case 'rate_limit_exceeded':
      return `Too many requests. Please wait ${error.retry_after}s.`
  }
}

export default function LimitToast({ error, onClose, onUpgrade }: LimitToastProps) {
  if (!error) return null

  const showUpgradeCta = error.error === 'user_limit_exceeded'

  return (
    <div
      style={{
        position: 'fixed',
        bottom: '24px',
        right: '24px',
        zIndex: 9999,
        maxWidth: '380px',
        width: 'calc(100vw - 48px)',
        background: '#1e293b',
        color: '#f1f5f9',
        borderRadius: '10px',
        padding: '16px 20px',
        boxShadow: '0 10px 40px rgba(0,0,0,0.3)',
        display: 'flex',
        flexDirection: 'column',
        gap: '12px',
        animation: 'slideUp 0.25s ease',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: '12px' }}>
        <p style={{ margin: 0, fontSize: '14px', lineHeight: 1.5, flex: 1 }}>{getMessage(error)}</p>
        <button
          onClick={onClose}
          aria-label="Dismiss"
          style={{
            background: 'none',
            border: 'none',
            color: '#94a3b8',
            cursor: 'pointer',
            fontSize: '18px',
            lineHeight: 1,
            padding: '0 2px',
            flexShrink: 0,
          }}
        >
          &times;
        </button>
      </div>

      {showUpgradeCta && onUpgrade && (
        <button
          onClick={() => {
            onUpgrade()
            onClose()
          }}
          style={{
            background: '#6366f1',
            color: '#fff',
            border: 'none',
            borderRadius: '6px',
            padding: '8px 14px',
            fontSize: '13px',
            fontWeight: 600,
            cursor: 'pointer',
            alignSelf: 'flex-start',
          }}
        >
          Upgrade to Pro
        </button>
      )}

      <style>{`
        @keyframes slideUp {
          from { transform: translateY(16px); opacity: 0; }
          to { transform: translateY(0); opacity: 1; }
        }
      `}</style>
    </div>
  )
}
