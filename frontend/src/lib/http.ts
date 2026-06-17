export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

export interface HttpOptions {
  method?: HttpMethod
  headers?: Record<string, string>
  body?: unknown
  signal?: AbortSignal
}

export interface HttpResponse<T = unknown> {
  status: number
  data: T
}

export async function request<T = unknown>(
  url: string,
  options: HttpOptions = {},
): Promise<HttpResponse<T>> {
  const { method = 'GET', headers = {}, body, signal } = options

  const init: RequestInit = {
    method,
    headers: { 'Content-Type': 'application/json', ...headers },
    signal,
  }

  if (body !== undefined) {
    init.body = JSON.stringify(body)
  }

  const response = await fetch(url, init)

  const contentType = response.headers.get('content-type') ?? ''
  let data: T
  if (contentType.includes('application/json')) {
    data = (await response.json()) as T
  } else {
    const text = await response.text()
    data = (text || null) as T
  }

  return { status: response.status, data }
}
