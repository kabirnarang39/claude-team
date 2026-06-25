import React, { useState } from 'react'
import { PaymentElement, useStripe, useElements } from '@stripe/react-stripe-js'
import { subscribe } from '../../api/billing'
import { LimitError, LimitErrorBody } from '../../api/client'

interface UpgradeFormProps {
  onSuccess: () => void
  onError: (message: string) => void
  onLimitError?: (limitBody: LimitErrorBody) => void
  startPolling: () => void
}

export default function UpgradeForm({ onSuccess, onError, onLimitError, startPolling }: UpgradeFormProps) {
  const stripe = useStripe()
  const elements = useElements()
  const [loading, setLoading] = useState(false)
  const [stripeError, setStripeError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    if (!stripe || !elements) {
      onError('Stripe has not loaded yet. Please try again.')
      return
    }

    setLoading(true)
    setStripeError(null)

    try {
      // Validate and submit the Payment Element
      const { error: submitError } = await elements.submit()
      if (submitError) {
        setStripeError(submitError.message ?? 'Payment failed. Please try again.')
        setLoading(false)
        return
      }

      // Create a PaymentMethod from the Payment Element
      const { paymentMethod, error: pmError } = await stripe.createPaymentMethod({
        elements,
      })

      if (pmError || !paymentMethod) {
        setStripeError(pmError?.message ?? 'Failed to create payment method.')
        setLoading(false)
        return
      }

      // Call backend subscribe API
      const result = await subscribe('pro', paymentMethod.id)

      if (result.status === 'pending') {
        // Start polling for active status
        startPolling()
        onSuccess()
      } else if (result.status === 'active') {
        onSuccess()
      } else {
        onError(`Unexpected subscription status: ${result.status}`)
      }
    } catch (err) {
      if (err instanceof LimitError && onLimitError) {
        onLimitError(err.limitBody)
      } else {
        const message = err instanceof Error ? err.message : 'An unexpected error occurred.'
        onError(message)
      }
    } finally {
      setLoading(false)
    }
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
      <h3 style={{ margin: '0 0 16px', fontSize: '18px', fontWeight: 700, color: '#1e293b' }}>
        Upgrade to Pro
      </h3>

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: '20px' }}>
          <PaymentElement
            options={{
              layout: 'tabs',
            }}
          />
        </div>

        {stripeError && (
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
            {stripeError}
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
          {loading ? 'Processing…' : 'Upgrade to Pro — $49/mo'}
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
