import React from 'react'
import { Invoice } from '../../api/billing'

interface InvoiceTableProps {
  invoices: Invoice[]
  loading: boolean
}

function formatAmount(cents: number, currency: string): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency.toUpperCase(),
    minimumFractionDigits: 2,
  }).format(cents / 100)
}

function formatDate(dateStr: string): string {
  return new Intl.DateTimeFormat('en-US', { year: 'numeric', month: 'short', day: 'numeric' }).format(
    new Date(dateStr)
  )
}

type StatusBadgeVariant = Invoice['status']

const STATUS_STYLES: Record<
  StatusBadgeVariant,
  { background: string; color: string; label: string }
> = {
  paid: { background: '#f0fdf4', color: '#166534', label: 'Paid' },
  open: { background: '#fffbeb', color: '#92400e', label: 'Open' },
  draft: { background: '#fffbeb', color: '#92400e', label: 'Draft' },
  uncollectible: { background: '#fef2f2', color: '#991b1b', label: 'Uncollectible' },
  void: { background: '#fef2f2', color: '#991b1b', label: 'Void' },
}

function StatusBadge({ status }: { status: StatusBadgeVariant }) {
  const style = STATUS_STYLES[status]
  return (
    <span
      style={{
        display: 'inline-block',
        padding: '2px 8px',
        borderRadius: '999px',
        fontSize: '12px',
        fontWeight: 600,
        background: style.background,
        color: style.color,
      }}
    >
      {style.label}
    </span>
  )
}

export default function InvoiceTable({ invoices, loading }: InvoiceTableProps) {
  if (loading) {
    return (
      <div style={{ padding: '24px', textAlign: 'center', color: '#94a3b8', fontSize: '14px' }}>
        Loading invoices…
      </div>
    )
  }

  if (invoices.length === 0) {
    return (
      <div
        style={{
          padding: '32px',
          textAlign: 'center',
          color: '#94a3b8',
          fontSize: '14px',
          border: '1px dashed #e2e8f0',
          borderRadius: '8px',
        }}
      >
        No invoices yet.
      </div>
    )
  }

  const tableStyle: React.CSSProperties = {
    width: '100%',
    borderCollapse: 'collapse',
    fontSize: '14px',
  }

  const thStyle: React.CSSProperties = {
    textAlign: 'left',
    padding: '10px 12px',
    borderBottom: '2px solid #e2e8f0',
    color: '#64748b',
    fontWeight: 600,
    fontSize: '12px',
    textTransform: 'uppercase',
    letterSpacing: '0.05em',
  }

  const tdStyle: React.CSSProperties = {
    padding: '12px',
    borderBottom: '1px solid #f1f5f9',
    color: '#334155',
    verticalAlign: 'middle',
  }

  return (
    <div style={{ overflowX: 'auto', borderRadius: '8px', border: '1px solid #e2e8f0' }}>
      <table style={tableStyle}>
        <thead>
          <tr>
            <th style={thStyle}>Date</th>
            <th style={thStyle}>Amount</th>
            <th style={thStyle}>Status</th>
            <th style={thStyle}>Invoice</th>
          </tr>
        </thead>
        <tbody>
          {invoices.map((invoice) => (
            <tr key={invoice.id} style={{ background: '#fff' }}>
              <td style={tdStyle}>{formatDate(invoice.period_start)}</td>
              <td style={{ ...tdStyle, fontWeight: 600 }}>
                {formatAmount(invoice.amount_cents, invoice.currency)}
              </td>
              <td style={tdStyle}>
                <StatusBadge status={invoice.status} />
              </td>
              <td style={tdStyle}>
                {invoice.invoice_pdf_url ? (
                  <a
                    href={invoice.invoice_pdf_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    style={{ color: '#6366f1', textDecoration: 'none', fontWeight: 500 }}
                  >
                    Download PDF
                  </a>
                ) : (
                  <span style={{ color: '#94a3b8' }}>—</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
