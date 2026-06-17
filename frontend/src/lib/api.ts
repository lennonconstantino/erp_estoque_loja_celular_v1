import { env } from '@/lib/env'
import { request } from '@/lib/http'
import { clearTokens, getAccessToken, getRefreshToken, setTokens } from '@/lib/auth'

const TIMEOUT_MS = 15_000

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly code: string,
    message: string,
    public readonly isNetworkError = false,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

function withTimeout<T>(promise: Promise<T>, ms: number): Promise<T> {
  return new Promise((resolve, reject) => {
    const id = setTimeout(
      () => reject(new ApiError(0, 'TIMEOUT', 'Tempo de requisição esgotado')),
      ms,
    )
    promise.then(
      (v) => { clearTimeout(id); resolve(v) },
      (e) => { clearTimeout(id); reject(e) },
    )
  })
}

let refreshing: Promise<void> | null = null

async function tryRefresh(): Promise<void> {
  const rt = getRefreshToken()
  if (!rt) {
    clearTokens()
    throw new ApiError(401, 'NO_REFRESH_TOKEN', 'Sessão expirada')
  }
  const res = await request<{ access_token: string; refresh_token: string }>(
    `${env.apiBaseUrl}/api/v1/auth/refresh`,
    { method: 'POST', body: { refresh_token: rt } },
  )
  if (res.status !== 200) {
    clearTokens()
    throw new ApiError(401, 'REFRESH_FAILED', 'Sessão expirada')
  }
  setTokens(res.data.access_token, res.data.refresh_token)
}

async function call<T>(
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
  path: string,
  body?: unknown,
  retry = true,
): Promise<T> {
  const token = getAccessToken()
  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  let res
  try {
    res = await withTimeout(
      request<T>(`${env.apiBaseUrl}${path}`, { method, headers, body }),
      TIMEOUT_MS,
    )
  } catch (err) {
    if (err instanceof ApiError) throw err
    throw new ApiError(0, 'NETWORK_ERROR', 'Erro de rede', true)
  }

  if (res.status === 401 && retry && getRefreshToken()) {
    if (!refreshing) refreshing = tryRefresh().finally(() => { refreshing = null })
    await refreshing
    return call(method, path, body, false)
  }

  if (res.status >= 400) {
    const errBody = res.data as { error?: { code?: string; message?: string } }
    throw new ApiError(
      res.status,
      errBody.error?.code ?? 'UNKNOWN',
      errBody.error?.message ?? 'Erro desconhecido',
    )
  }

  return res.data
}

export const api = {
  get:    <T>(path: string)                  => call<T>('GET',    path),
  post:   <T>(path: string, body?: unknown)  => call<T>('POST',   path, body),
  put:    <T>(path: string, body?: unknown)  => call<T>('PUT',    path, body),
  patch:  <T>(path: string, body?: unknown)  => call<T>('PATCH',  path, body),
  delete: <T>(path: string)                  => call<T>('DELETE', path),
}
