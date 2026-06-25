import React from 'react'
import { Usage } from '../../api/billing'

interface UsageBarProps {
  usage: Usage | null
}

function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'short', day: 'numeric' }).format(
    new Date(dateStr)
  )
}

function getBarColor(pct: number): string {
  if (pct >= 90) return '#ef4444'
  if (pct >= 70) return '#f59e0b'
  return '#22c55e'
}

export default function UsageBar({ usage }: UsageBarProps) {
  if (!usage) {
    return (
      <div style={{ padding: '16px', background: '#f8fafc', borderRadius: '8px', maxWidth: '480px' }}>
        <p style={{ color: '#94a3b8', margin: 0, fontSize: '14px' }}>Loading usage…</p>
      </div>
    )
  }

  if (usage.limit === null) {
    return (
      <div style={{ padding: '16px', background: '#f8fafc', borderRadius: '8px', maxWidth: '480px' }}>
        <p style={{ color: '#64748b', margin: '0 0 4px', fontSize: '14px', fontWeight: 600 }}>API Usage</p>
        <p style={{ color: '#475569', margin: 0, fontSize: '14px' }}>No API quota — unlimited calls.</p>
        <p style={{ color: '#94a3b8', margin: '4px 0 0', fontSize: '12px' }}>
          Period ends {formatDate(usage.period_end)}
        </p>
      </div>
    )
  }

  const pct = Math.min(100, (usage.current_period_calls / usage.limit) * 100)
  const barColor = getBarColor(pct)

  return (
    <div style={{ padding: '16px', background: '#f8fafc', borderRadius: '8px', maxWidth: '480px' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: '8px' }}>
        <span style={{ fontSize: '14px', fontWeight: 600, color: '#475569' }}>API Usage</span>
        <span style={{ fontSize: '13px', color: '#64748b' }}>
          {usage.current_period_calls.toLocaleString()} / {usage.limit.toLocaleString()} calls
        </span>
      </div>

      <div
        style={{
          background: '#e2e8f0',
          borderRadius: '999px',
          height: '8px',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: `${pct}%`,
            height: '100%',
            background: barColor,
            borderRadius: '999px',
            transition: 'width 0.4s ease',
          }}
        />
      </div>

      <p style={{ margin: '6px 0 0', fontSize: '12px', color: '#94a3b8' }}>
        Period ends {formatDate(usage.period_end)}
      </p>
    </div>
  )
}
