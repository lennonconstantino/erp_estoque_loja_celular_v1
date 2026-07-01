const REFRESH_KEY = 'erp_refresh_token'
const PERMS_KEY = 'erp_perms'
const USER_ID_KEY = 'erp_user_id'

let accessToken: string | null = null

export function getAccessToken(): string | null {
  return accessToken
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY)
}

export function setTokens(access: string, refresh: string): void {
  accessToken = access
  localStorage.setItem(REFRESH_KEY, refresh)
  // Persiste claims do token (perms + id do usuário) para gating de UI que
  // sobreviva a reload — o access token vive só em memória.
  const { sub, perms } = decodeClaims(access)
  localStorage.setItem(PERMS_KEY, JSON.stringify(perms))
  if (sub) localStorage.setItem(USER_ID_KEY, sub)
  else localStorage.removeItem(USER_ID_KEY)
}

export function clearTokens(): void {
  accessToken = null
  localStorage.removeItem(REFRESH_KEY)
  localStorage.removeItem(PERMS_KEY)
  localStorage.removeItem(USER_ID_KEY)
}

export function isAuthenticated(): boolean {
  return accessToken !== null
}

/**
 * Permissões (claim `perms`) do usuário autenticado, lidas do último token.
 *
 * ATENÇÃO: uso **exclusivamente cosmético** — mostrar/ocultar menu e telas. A
 * autorização real é feita pelo backend, que valida a assinatura do JWT e a
 * permissão exigida (`iam:admin`, etc.) em toda requisição protegida. Nunca
 * trate isto como fronteira de segurança.
 */
export function getPerms(): string[] {
  try {
    const raw = localStorage.getItem(PERMS_KEY)
    const parsed = raw ? (JSON.parse(raw) as unknown) : null
    return Array.isArray(parsed) ? (parsed as string[]) : []
  } catch {
    return []
  }
}

export function hasPerm(perm: string): boolean {
  return getPerms().includes(perm)
}

/** Id (claim `sub`) do usuário autenticado — usado só para guard-rails de UI. */
export function getUserId(): string | null {
  return localStorage.getItem(USER_ID_KEY)
}

/** Decodifica o payload de um JWT e extrai `sub`/`perms` (sem validar assinatura). */
function decodeClaims(accessToken: string): { sub: string | null; perms: string[] } {
  try {
    const payload = accessToken.split('.')[1]
    if (!payload) return { sub: null, perms: [] }
    const base64url = payload.replace(/-/g, '+').replace(/_/g, '/')
    // JWT base64url não traz padding; `atob` pode rejeitar sem ele.
    const base64 = base64url + '='.repeat((4 - (base64url.length % 4)) % 4)
    const json = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + c.charCodeAt(0).toString(16).padStart(2, '0'))
        .join(''),
    )
    const claims = JSON.parse(json) as { sub?: unknown; perms?: unknown }
    return {
      sub: typeof claims.sub === 'string' ? claims.sub : null,
      perms: Array.isArray(claims.perms) ? (claims.perms as string[]) : [],
    }
  } catch {
    return { sub: null, perms: [] }
  }
}
