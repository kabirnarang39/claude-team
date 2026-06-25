// Base API client
// - Reads JWT from localStorage key "token"
// - Attaches Authorization: Bearer {token} header
// - For POST requests: generates UUID v4 Idempotency-Key header
// - Handles 401 (clear token, redirect to /login), 403 (throw ForbiddenError), 429 (throw LimitError)

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    public body: unknown
  ) {
    super(code)
    this.name = 'ApiError'
  }
}

export class ForbiddenError extends ApiError {
  constructor(body: unknown) {
    super(403, 'forbidden', body)
    this.name = 'ForbiddenError'
  }
}

export interface LimitExceededBody {
  error: 'user_limit_exceeded'
  limit: number
  current: number
  plan: string
}

export interface QuotaExceededBody {
  error: 'api_quota_exceeded'
  limit: number
  current: number
  plan: string
  reset_at: string
}

export interface RateLimitBody {
  error: 'rate_limit_exceeded'
  retry_after: number
}

export type LimitErrorBody = LimitExceededBody | QuotaExceededBody | RateLimitBody

export class LimitError extends ApiError {
  constructor(public limitBody: LimitErrorBody) {
    super(429, limitBody.error, limitBody)
    this.name = 'LimitError'
  }
}

export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('token')

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string> | undefined),
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  const method = options?.method?.toUpperCase() ?? 'GET'
  if (method === 'POST' || method === 'PUT' || method === 'PATCH') {
    headers['Idempotency-Key'] = crypto.randomUUID()
  }

  const response = await fetch(path, {
    ...options,
    headers,
  })

  if (response.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new ApiError(401, 'unauthorized', null)
  }

  if (response.status === 403) {
    const body: unknown = await response.json().catch(() => null)
    throw new ForbiddenError(body)
  }

  if (response.status === 429) {
    const body: unknown = await response.json().catch(() => null)
    const limitBody = body as LimitErrorBody
    throw new LimitError(limitBody)
  }

  if (!response.ok) {
    const body: unknown = await response.json().catch(() => null)
    const code = (body as { error?: string } | null)?.error ?? `http_${response.status}`
    throw new ApiError(response.status, code, body)
  }

  // 204 No Content
  if (response.status === 204) {
    return undefined as unknown as T
  }

  return response.json() as Promise<T>
}
