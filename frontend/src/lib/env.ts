function required(key: string): string {
  const value = import.meta.env[key] as string | undefined
  if (!value) throw new Error(`[env] Variável obrigatória ausente: ${key}`)
  return value
}

export const env = {
  apiBaseUrl: required('VITE_API_BASE_URL'),
} as const
