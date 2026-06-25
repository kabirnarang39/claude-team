import React, { useState } from 'react'
import { PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js'
import { getPortalUrl } from '../../api/billing'

interface PaymentMethodFormProps {
  onSuccess: () => void
}

export default function PaymentMethodForm({ onSuccess }: PaymentMethodFormProps) {
  const stripe = useStripe()
  const elements = useElements()
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const handlePortalRedirect = async () => {
    setLoading(true)
    setError(null)
    try {
      const { url } = await getPortalUrl()
      window.location.href = url
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not open billing portal.')
      setLoading(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    if (!stripe || !elements) {
      setError('Stripe has not loaded yet. Please try again.')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const { error: submitError } = await elements.submit()
      if (submitError) {
        setError(submitError.message ?? 'Submission failed.')
        setLoading(false)
        return
      }

      const { error: confirmError } = await stripe.confirmSetup({
        elements,
        confirmParams: {
          return_url: `${window.location.origin}/billing`,
        },
      })

      if (confirmError) {
        setError(confirmError.message ?? 'Setup failed.')
      } else {
        setSuccess(true)
        onSuccess()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An unexpected error occurred.')
    } finally {
      setLoading(false)
    }
  }

  if (success) {
    return (
      <div
        style={{
          padding: '16px',
          background: '#f0fdf4',
          border: '1px solid #bbf7d0',
          borderRadius: '8px',
          color: '#166534',
          fontSize: '14px',
          fontWeight: 500,
        }}
      >
        Payment method updated successfully.
      </div>
    )
  }

  return (
    <div
      style={{
        border: '1px solid #e2e8f0',
        borderRadius: '12px',
        padding: '24px',
        maxWidth: '480px',
        background: '#fff',
      }}
    >
      <h3 style={{ margin: '0 0 8px', fontSize: '18px', fontWeight: 700, color: '#1e293b' }}>
        Update Payment Method
      </h3>
      <p style={{ margin: '0 0 16px', fontSize: '14px', color: '#64748b' }}>
        Or{' '}
        <button
          onClick={handlePortalRedirect}
          style={{
            background: 'none',
            border: 'none',
            color: '#6366f1',
            cursor: 'pointer',
            fontSize: '14px',
            padding: 0,
            textDecoration: 'underline',
          }}
        >
          manage via Stripe portal
        </button>
      </p>

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '20px' }}>
          <PaymentElement />
        </div>

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

        <button
          type="submit"
          disabled={loading || !stripe || !elements}
          style={{
            width: '100%',
            background: loading ? '#a5b4fc' : '#6366f1',
            color: '#fff',
            border: 'none',
            borderRadius: '8px',
            padding: '12px 20px',
            fontSize: '15px',
            fontWeight: 600,
            cursor: loading ? 'not-allowed' : 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px',
          }}
        >
          {loading && (
            <span
              style={{
                width: '16px',
                height: '16px',
                border: '2px solid rgba(255,255,255,0.4)',
                borderTopColor: '#fff',
                borderRadius: '50%',
                display: 'inline-block',
                animation: 'spin 0.7s linear infinite',
              }}
            />
          )}
          {loading ? 'Saving…' : 'Save Payment Method'}
        </button>
      </form>

      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  )
}
