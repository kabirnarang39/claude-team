import { apiFetch } from './client'

export interface Subscription {
  id: string
  workspace_id: string
  plan: 'free' | 'pro' | 'enterprise'
  status: 'active' | 'pending' | 'past_due' | 'cancelled' | 'trialing'
  stripe_subscription_id: string | null
  stripe_customer_id: string | null
  current_period_start: string | null
  current_period_end: string | null
  cancel_at_period_end: boolean
  cancel_at: string | null
  user_limit: number
  api_call_limit: number | null
}

export interface Usage {
  current_period_calls: number
  limit: number | null
  period_start: string
  period_end: string
}

export interface Invoice {
  id: string
  stripe_invoice_id: string
  amount_cents: number
  currency: string
  status: 'paid' | 'open' | 'draft' | 'uncollectible' | 'void'
  period_start: string
  period_end: string
  paid_at: string | null
  invoice_pdf_url: string | null
}

export interface PaginatedInvoices {
  invoices: Invoice[]
  total: number
  page: number
  per_page: number
}

export function subscribe(
  planId: string,
  paymentMethodId: string
): Promise<{ status: string; stripe_subscription_id: string }> {
  return apiFetch('/api/billing/subscribe', {
    method: 'POST',
    body: JSON.stringify({ plan_id: planId, payment_method_id: paymentMethodId }),
  })
}

export function downgrade(): Promise<{ cancel_at: string; status: string }> {
  return apiFetch('/api/billing/downgrade', { method: 'POST', body: '{}' })
}

export function cancel(immediate = false): Promise<{ status: string }> {
  const qs = immediate ? '?immediate=true' : ''
  return apiFetch(`/api/billing/cancel${qs}`, { method: 'POST', body: '{}' })
}

export function getSubscription(): Promise<Subscription> {
  return apiFetch('/api/billing/subscription')
}

export function getUsage(): Promise<Usage> {
  return apiFetch('/api/billing/usage')
}

export function getInvoices(page = 1, perPage = 20): Promise<PaginatedInvoices> {
  return apiFetch(`/api/billing/invoices?page=${page}&per_page=${perPage}`)
}

export function getPortalUrl(): Promise<{ url: string }> {
  return apiFetch('/api/billing/portal', { method: 'POST', body: '{}' })
}
