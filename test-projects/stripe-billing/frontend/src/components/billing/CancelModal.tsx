import React, { useState } from 'react'
import { cancel } from '../../api/billing'

interface CancelModalProps {
  isOpen: boolean
  periodEndDate: string | null
  onClose: () => void
  onCancelled: () => void
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return 'the end of your billing period'
  return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'long', day: 'numeric' }).format(
    new Date(dateStr)
  )
}

export default function CancelModal({ isOpen, periodEndDate, onClose, onCancelled }: CancelModalProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  if (!isOpen) return null

  const handleCancel = async () => {
    setLoading(true)
    setError(null)
    try {
      await cancel(false)
      onCancelled()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cancel subscription.')
    } finally {
      setLoading(false)
    }
  }

  const overlayStyle: React.CSSProperties = {
    position: 'fixed',
    inset: 0,
    background: 'rgba(15,23,42,0.5)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    zIndex: 1000,
  }

  const modalStyle: React.CSSProperties = {
    background: '#fff',
    borderRadius: '12px',
    padding: '28px',
    maxWidth: '420px',
    width: '90%',
    boxShadow: '0 20px 60px rgba(0,0,0,0.2)',
  }

  return (
    <div style={overlayStyle} onClick={onClose}>
      <div style={modalStyle} onClick={(e) => e.stopPropagation()}>
        <h2 style={{ margin: '0 0 12px', fontSize: '20px', fontWeight: 700, color: '#1e293b' }}>
          Cancel Subscription?
        </h2>
        <p style={{ margin: '0 0 20px', fontSize: '15px', color: '#475569', lineHeight: 1.6 }}>
          Your Pro subscription will end on{' '}
          <strong style={{ color: '#1e293b' }}>{formatDate(periodEndDate)}</strong>. You'll keep Pro access
          until then.
        </p>

        {error && (
          <p
            style={{
              color: '#ef4444',
              fontSize: '14px',
              margin: '0 0 16px',
              padding: '10px 12px',
              background: '#fef2f2',
              borderRadius: '6px',
              border: '1px solid #fecaca',
            }}
          >
            {error}
          </p>
        )}

        <div style={{ display: 'flex', gap: '12px', justifyContent: 'flex-end' }}>
          <button
            onClick={onClose}
            disabled={loading}
            style={{
              background: '#f1f5f9',
              color: '#334155',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              padding: '10px 18px',
              fontSize: '14px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Keep Subscription
          </button>
          <button
            onClick={handleCancel}
            disabled={loading}
            style={{
              background: loading ? '#fca5a5' : '#ef4444',
              color: '#fff',
              border: 'none',
              borderRadius: '8px',
              padding: '10px 18px',
              fontSize: '14px',
              fontWeight: 600,
              cursor: loading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            {loading && (
              <span
                style={{
                  width: '14px',
                  height: '14px',
                  border: '2px solid rgba(255,255,255,0.4)',
                  borderTopColor: '#fff',
                  borderRadius: '50%',
                  display: 'inline-block',
                  animation: 'spin 0.7s linear infinite',
                }}
              />
            )}
            {loading ? 'Cancelling…' : 'Cancel Subscription'}
          </button>
        </div>
      </div>

      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}
