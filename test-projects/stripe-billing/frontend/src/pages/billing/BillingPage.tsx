import React, { useState } from 'react'
import { useSubscription } from '../../hooks/useSubscription'
import { useUsage } from '../../hooks/useUsage'
import { useInvoices } from '../../hooks/useInvoices'
import PlanCard from '../../components/billing/PlanCard'
import UsageBar from '../../components/billing/UsageBar'
import UpgradeForm from '../../components/billing/UpgradeForm'
import CancelModal from '../../components/billing/CancelModal'
import InvoiceTable from '../../components/billing/InvoiceTable'
import LimitToast from '../../components/billing/LimitToast'
import { getPortalUrl } from '../../api/billing'
import { LimitError, LimitErrorBody } from '../../api/client'

export default function BillingPage() {
  const { subscription, loading: subLoading, refetch: refetchSub, startPolling } = useSubscription()
  const { usage, loading: usageLoading } = useUsage()
  const { invoices, loading: invoicesLoading, total, page, perPage, setPage } = useInvoices()

  const [showUpgradeForm, setShowUpgradeForm] = useState(false)
  const [showCancelModal, setShowCancelModal] = useState(false)
  const [limitError, setLimitError] = useState<LimitErrorBody | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const isPro = subscription?.plan === 'pro' || subscription?.plan === 'enterprise'
  const isCancellingAlready = subscription?.cancel_at_period_end === true

  const handleUpgrade = () => {
    setShowUpgradeForm(true)
  }

  const handleManage = async () => {
    try {
      const { url } = await getPortalUrl()
      window.location.href = url
    } catch (err) {
      if (err instanceof LimitError) {
        setLimitError(err.limitBody)
      }
    }
  }

  const handleUpgradeSuccess = () => {
    setSuccessMessage('Upgraded to Pro! Your subscription is being activated.')
    startPolling()
    setTimeout(() => {
      setShowUpgradeForm(false)
      refetchSub()
      setSuccessMessage(null)
    }, 3000)
  }

  const handleUpgradeError = (message: string) => {
    // Stripe errors are shown inline by UpgradeForm.
    // Non-limit API errors surface here as a generic notice.
    console.error('Upgrade error:', message)
  }

  const handleCancelled = () => {
    setShowCancelModal(false)
    refetchSub()
    setSuccessMessage('Subscription set to cancel at period end.')
    setTimeout(() => setSuccessMessage(null), 4000)
  }

  const totalPages = Math.ceil(total / perPage)

  if (subLoading) {
    return (
      <div style={{ padding: '40px', color: '#64748b', fontSize: '16px' }}>
        Loading billing information…
      </div>
    )
  }

  return (
    <div
      style={{
        maxWidth: '640px',
        margin: '0 auto',
        padding: '40px 24px',
        fontFamily:
          '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      }}
    >
      <h1 style={{ fontSize: '28px', fontWeight: 800, color: '#0f172a', margin: '0 0 32px' }}>Billing</h1>

      {successMessage && (
        <div
          style={{
            padding: '12px 16px',
            background: '#f0fdf4',
            border: '1px solid #bbf7d0',
            borderRadius: '8px',
            color: '#166534',
            fontSize: '14px',
            fontWeight: 500,
            marginBottom: '24px',
          }}
        >
          {successMessage}
        </div>
      )}

      {/* 1. Plan Card */}
      <section style={{ marginBottom: '24px' }}>
        <PlanCard
          subscription={subscription}
          onUpgrade={handleUpgrade}
          onManage={handleManage}
        />
      </section>

      {/* 2. Usage Bar */}
      {!usageLoading && usage && (
        <section style={{ marginBottom: '24px' }}>
          <UsageBar usage={usage} />
        </section>
      )}

      {/* 3. Upgrade Form (shown when free or user clicked upgrade) */}
      {(!isPro || showUpgradeForm) && (
        <section style={{ marginBottom: '24px' }}>
          <UpgradeForm
            onSuccess={handleUpgradeSuccess}
            onError={handleUpgradeError}
            onLimitError={setLimitError}
            startPolling={startPolling}
          />
        </section>
      )}

      {/* 4. Cancel button (shown only for active Pro, not already cancelling) */}
      {isPro && !isCancellingAlready && (
        <section style={{ marginBottom: '24px' }}>
          <button
            onClick={() => setShowCancelModal(true)}
            style={{
              background: 'none',
              border: '1px solid #fecaca',
              borderRadius: '8px',
              padding: '10px 18px',
              fontSize: '14px',
              fontWeight: 500,
              color: '#ef4444',
              cursor: 'pointer',
            }}
          >
            Cancel Subscription
          </button>
        </section>
      )}

      {/* 5. Invoice Table */}
      <section>
        <h2 style={{ fontSize: '18px', fontWeight: 700, color: '#1e293b', margin: '0 0 16px' }}>
          Invoices
        </h2>
        <InvoiceTable invoices={invoices} loading={invoicesLoading} />

        {totalPages > 1 && (
          <div
            style={{
              display: 'flex',
              justifyContent: 'center',
              alignItems: 'center',
              gap: '12px',
              marginTop: '16px',
            }}
          >
            <button
              onClick={() => setPage(page - 1)}
              disabled={page <= 1}
              style={{
                background: '#f1f5f9',
                border: '1px solid #e2e8f0',
                borderRadius: '6px',
                padding: '6px 14px',
                fontSize: '14px',
                cursor: page <= 1 ? 'not-allowed' : 'pointer',
                color: page <= 1 ? '#94a3b8' : '#334155',
              }}
            >
              Previous
            </button>
            <span style={{ fontSize: '14px', color: '#64748b' }}>
              Page {page} of {totalPages}
            </span>
            <button
              onClick={() => setPage(page + 1)}
              disabled={page >= totalPages}
              style={{
                background: '#f1f5f9',
                border: '1px solid #e2e8f0',
                borderRadius: '6px',
                padding: '6px 14px',
                fontSize: '14px',
                cursor: page >= totalPages ? 'not-allowed' : 'pointer',
                color: page >= totalPages ? '#94a3b8' : '#334155',
              }}
            >
              Next
            </button>
          </div>
        )}
      </section>

      {/* Cancel Modal */}
      <CancelModal
        isOpen={showCancelModal}
        periodEndDate={subscription?.current_period_end ?? null}
        onClose={() => setShowCancelModal(false)}
        onCancelled={handleCancelled}
      />

      {/* Limit Toast */}
      <LimitToast
        error={limitError}
        onClose={() => setLimitError(null)}
        onUpgrade={handleUpgrade}
      />
    </div>
  )
}
