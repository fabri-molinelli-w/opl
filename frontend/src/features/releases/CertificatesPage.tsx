import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getListCertificatesQueryKey,
  useCheckCertificate,
  useCreateCertificate,
  useDeleteCertificate,
  useListCertificates,
} from '@/api/gen'
import type { CertificateStatus } from '@/api/gen/model'
import { hasRole } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { fieldErrors, formErrorMessage } from '@/features/inventory/formErrors'
import { formatDateTime } from '@/features/incidents/labels'

const CERT_STYLES: Record<CertificateStatus, string> = {
  ok: 'bg-emerald-100 text-emerald-800',
  warning: 'bg-amber-100 text-amber-800',
  critical: 'bg-red-100 text-red-800',
  unknown: 'bg-slate-200 text-slate-700',
}

const CERT_LABELS: Record<CertificateStatus, string> = {
  ok: 'OK',
  warning: 'Por vencer (≤ 30 días)',
  critical: 'Crítico (≤ 7 días)',
  unknown: 'Sin datos',
}

/** CU-06: certificados SSL vigilados, ordenados por días restantes. */
export const CertificatesPage = () => {
  const { user } = useAuth()
  const canWrite = hasRole(user?.role, 'sysadmin')
  const queryClient = useQueryClient()
  const [status, setStatus] = useState('')
  const list = useListCertificates(status === '' ? undefined : { status: status as CertificateStatus })
  const items = list.data?.status === 200 ? list.data.data : []
  const create = useCreateCertificate()
  const check = useCheckCertificate()
  const remove = useDeleteCertificate()
  const [domain, setDomain] = useState('')
  const [port, setPort] = useState('443')
  const [error, setError] = useState('')

  const refresh = () => queryClient.invalidateQueries({ queryKey: getListCertificatesQueryKey() })

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await create.mutateAsync({ data: { domain, port: Number(port) || 443 } })
      setDomain('')
      await refresh()
    } catch (cause) {
      const fields = fieldErrors(cause)
      setError(fields.domain ?? formErrorMessage(cause, 'No se pudo agregar el certificado'))
    }
  }

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-semibold">Certificados SSL</h1>

      {canWrite && (
        <form
          onSubmit={(e) => void submit(e)}
          className="flex flex-wrap items-end gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
        >
          <label className="text-sm">
            Dominio
            <input
              required
              value={domain}
              onChange={(e) => setDomain(e.target.value)}
              placeholder="lab-tls.lab.local"
              className="mt-1 block rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
          <label className="text-sm">
            Puerto
            <input
              type="number"
              min={1}
              max={65535}
              value={port}
              onChange={(e) => setPort(e.target.value)}
              className="mt-1 block w-24 rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
          <button
            type="submit"
            disabled={create.isPending}
            className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
          >
            {create.isPending ? 'Chequeando…' : 'Agregar y chequear'}
          </button>
          {error !== '' && <p className="w-full text-sm text-red-600">{error}</p>}
        </form>
      )}

      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <select
          aria-label="Estado"
          value={status}
          onChange={(e) => setStatus(e.target.value)}
          className="mb-4 rounded-md border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">Todos los estados</option>
          {(Object.keys(CERT_LABELS) as CertificateStatus[]).map((s) => (
            <option key={s} value={s}>
              {CERT_LABELS[s]}
            </option>
          ))}
        </select>
        <table className="w-full text-left text-sm">
          <thead className="text-slate-500">
            <tr>
              <th className="py-2">Dominio</th>
              <th>Estado</th>
              <th>Días restantes</th>
              <th>Vence</th>
              <th>Emisor</th>
              <th>Último chequeo</th>
              {canWrite && <th />}
            </tr>
          </thead>
          <tbody>
            {items.length === 0 && (
              <tr>
                <td colSpan={7} className="py-6 text-center text-slate-500">
                  {list.isLoading ? 'Cargando…' : 'No hay certificados vigilados.'}
                </td>
              </tr>
            )}
            {items.map((c) => (
              <tr key={c.id} className="border-t border-slate-100">
                <td className="py-2 font-medium">
                  {c.domain}:{c.port}
                  {c.error !== undefined && <span className="block text-xs font-normal text-red-600">{c.error}</span>}
                </td>
                <td>
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${CERT_STYLES[c.status]}`}>
                    {CERT_LABELS[c.status]}
                  </span>
                </td>
                <td>{c.daysLeft ?? '—'}</td>
                <td>{c.notAfter !== undefined ? formatDateTime(c.notAfter) : '—'}</td>
                <td>{c.issuer || '—'}</td>
                <td>{c.lastCheckedAt !== undefined ? formatDateTime(c.lastCheckedAt) : '—'}</td>
                {canWrite && (
                  <td className="space-x-2 text-right">
                    <button
                      type="button"
                      onClick={() => void check.mutateAsync({ id: c.id }).then(refresh)}
                      className="text-sky-700 hover:underline"
                    >
                      Chequear
                    </button>
                    <button
                      type="button"
                      onClick={() => void remove.mutateAsync({ id: c.id }).then(refresh)}
                      className="text-red-700 hover:underline"
                    >
                      Quitar
                    </button>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
