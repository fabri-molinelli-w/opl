import type { Role } from '@/api/gen/model'

export const ROLE_RANK: Record<Role, number> = {
  soporte: 1,
  sysadmin: 2,
  admin: 3,
}

export const ROLE_LABELS: Record<Role, string> = {
  admin: 'Administrador',
  sysadmin: 'SysAdmin',
  soporte: 'Soporte',
}

export const ALL_ROLES: readonly Role[] = ['admin', 'sysadmin', 'soporte']

export const hasRole = (role: Role | null | undefined, min: Role): boolean =>
  role !== null && role !== undefined && ROLE_RANK[role] >= ROLE_RANK[min]
