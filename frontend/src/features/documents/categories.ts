import type { DocumentCategory } from '@/api/gen/model'

export const CATEGORIES: DocumentCategory[] = [
  'procedimiento',
  'runbook',
  'guia',
  'configuracion',
  'referencia',
]

export const CATEGORY_LABELS: Record<DocumentCategory, string> = {
  procedimiento: 'Procedimiento',
  runbook: 'Runbook',
  guia: 'Guía',
  configuracion: 'Configuración',
  referencia: 'Referencia',
}

export const formatDate = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', {
    dateStyle: 'short',
    timeStyle: 'short',
  }).format(new Date(value))
