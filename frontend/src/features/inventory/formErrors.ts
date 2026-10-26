import { isApiError } from '@/api/http'

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

export const fieldErrors = (cause: unknown): Record<string, string> => {
  if (!isApiError(cause) || !isRecord(cause.data)) {
    return {}
  }
  const details = cause.data.details
  if (!Array.isArray(details)) {
    return {}
  }
  const errors: Record<string, string> = {}
  for (const detail of details) {
    if (!isRecord(detail)) continue
    const field = detail.field
    const issue = detail.issue
    if (typeof field === 'string' && typeof issue === 'string') {
      errors[field] = issue
    }
  }
  return errors
}

export const formErrorMessage = (
  cause: unknown,
  fallback: string,
): string => (isApiError(cause) ? cause.message : fallback)
