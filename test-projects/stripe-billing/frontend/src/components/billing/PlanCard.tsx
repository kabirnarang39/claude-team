import React from 'react'
import { Subscription } from '../../api/billing'

interface PlanCardProps {
  subscription: Subscription | null
  onUpgrade: () => void
  onManage: () => void
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '—'
  return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'long', day: 'numeric' }).format(
    new Date(dateStr)
  )
}

const PLAN_DETAILS: Record<string, { label: string; price: string; description: string }> = {
  free: { label: 'Free', price: '$0/month', description: 'Basic access for individuals' },
  pro: { label: 'Pro', price: '$49/month', description: 'Full access for growing teams' },
  enterprise: { label: 'Enterprise', price: 'Custom', description: 'Unlimited scale for large orgs' },
}

export default function PlanCard({ subscription, onUpgrade, onManage }: PlanCardProps) {
  const plan = subscription?.plan ?? 'free'
  const details = PLAN_DETAILS[plan] ?? PLAN_DETAILS['free']

  const cardStyle: React.CSSProperties = {
    border: '2px solid',
    borderColor: plan === 'pro' ? '#6366f1' : plan === 'enterprise' ? '#f59e0b' : '#e2e8f0',
    borderRadius: '12px',
    padding: '24px',
    background: '#fff',
    maxWidth: '480px',
    boxShadow: plan !== 'free' ? '0 4px 16px rgba(99,102,241,0.08)' : 'none',
  }

  const badgeStyle: React.CSSProperties = {
    display: 'inline-block',
    padding: '2px 10px',
    borderRadius: '999px',
    fontSize: '12px',
    fontWeight: 700,
    marginBottom: '8px',
    background: plan === 'pro' ? '#6366f1' : plan === 'enterprise' ? '#f59e0b' : '#e2e8f0',
    color: plan === 'free' ? '#475569' : '#fff',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  }

  return (
    <div style={cardStyle}>
      <span style={badgeStyle}>{details.label}</span>
      <h2 style={{ margin: '0 0 4px', fontSize: '24px', fontWeight: 700, color: '#1e293b' }}>
        {details.price}
      </h2>
      <p style={{ margin: '0 0 16px', color: '#64748b', fontSize: '14px' }}>{details.description}</p>

      <div style={{ display: 'flex', gap: '24px', marginBottom: '16px', fontSize: '14px' }}>
        <div>
          <span style={{ color: '#94a3b8', display: 'block', marginBottom: '2px' }}>Users</span>
          <span style={{ fontWeight: 600, color: '#1e293b' }}>
            {subscription ? `Up to ${subscription.user_limit}` : '1'}
          </span>
        </div>
        <div>
          <span style={{ color: '#94a3b8', display: 'block', marginBottom: '2px' }}>API Quota</span>
          <span style={{ fontWeight: 600, color: '#1e293b' }}>
            {subscription?.api_call_limit == null
              ? 'Unlimited'
              : `${subscription.api_call_limit.toLocaleString()} calls/mo`}
          </span>
        </div>
      </div>

      {plan === 'pro' && subscription?.current_period_start && (
        <p style={{ fontSize: '13px', color: '#64748b', margin: '0 0 8px' }}>
          Billing period:{' '}
          <strong>
            {formatDate(subscription.current_period_start)} – {formatDate(subscription.current_period_end)}
          </strong>
        </p>
      )}

      {subscription?.cancel_at_period_end && subscription.cancel_at && (
        <p style={{ fontSize: '13px', color: '#ef4444', margin: '0 0 12px', fontWeight: 500 }}>
          Cancels on {formatDate(subscription.cancel_at)} — you retain Pro access until then.
        </p>
      )}

      <div style={{ marginTop: '16px' }}>
        {plan === 'free' ? (
          <button
            onClick={onUpgrade}
            style={{
              background: '#6366f1',
              color: '#fff',
              border: 'none',
              borderRadius: '8px',
              padding: '10px 20px',
              fontSize: '14px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Upgrade to Pro
          </button>
        ) : (
          <button
            onClick={onManage}
            style={{
              background: '#f1f5f9',
              color: '#334155',
              border: '1px solid #e2e8f0',
              borderRadius: '8px',
              padding: '10px 20px',
              fontSize: '14px',
              fontWeight: 600,
              cursor: 'pointer',
            }}
          >
            Manage Subscription
          </button>
        )}
      </div>
    </div>
  )
}
