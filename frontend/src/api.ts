const BASE = '/api/v1'
export const DEMO_NAMESPACE = 'rpg-demo'

export type Difficulty = 'easy' | 'normal' | 'hard'

export interface ObjectMeta {
  name: string
  namespace?: string
  resourceVersion?: string
  generation?: number
  [field: string]: unknown
}

export interface KubernetesResource {
  apiVersion?: string
  kind?: string
  metadata: ObjectMeta
  [field: string]: unknown
}

export interface KroCondition {
  type: string
  status: string
  reason?: string
  message?: string
  observedGeneration?: number
}

export interface DungeonStatus {
  livingMonsters?: number | string
  bossState?: string
  bossPhase?: string
  maxHeroHP?: number | string
  maxMonsterHP?: number | string
  maxBossHP?: number | string
  victory?: boolean | string
  treasureState?: string
  observedGeneration?: number
  conditions?: KroCondition[]
}

export interface Dungeon extends KubernetesResource {
  spec: {
    monsters: number
    difficulty: Difficulty
    heroClass: 'warrior'
    attackSeq: number
    game: {
      heroHP: number
      bossHP: number
      monsterHP: number[]
      treasureOpened: 0 | 1
    }
  }
  status?: DungeonStatus
}

export interface AuthUser {
  login: string
  avatarUrl: string
}

export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')
  if (options.body) headers.set('Content-Type', 'application/json')
  const response = await fetch(`${BASE}${path}`, {
    credentials: 'same-origin',
    cache: 'no-store',
    ...options,
    signal: options.signal
      ? AbortSignal.any([options.signal, AbortSignal.timeout(15_000)])
      : AbortSignal.timeout(15_000),
    headers,
  })
  if (!response.ok) {
    let message = await response.text()
    try {
      const body: { error?: string; message?: string } = JSON.parse(message)
      message = body.error || body.message || message
    } catch {
      // Some ingress errors are plain text rather than API JSON.
    }
    throw new ApiError(response.status, message || `HTTP ${response.status}`)
  }
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

function dungeonPath(name: string) {
  return `/dungeons/${DEMO_NAMESPACE}/${encodeURIComponent(name)}`
}

export async function getMe(signal?: AbortSignal): Promise<AuthUser | null> {
  try {
    return await request<AuthUser>('/auth/me', { signal })
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) return null
    throw error
  }
}

export const login = () => request<AuthUser>('/auth/login', { method: 'POST' })
export const logout = () => request<void>('/auth/logout', { method: 'POST' })
export const listDungeons = (signal?: AbortSignal) =>
  request<Dungeon[]>('/dungeons', { signal })
export const getDungeon = (name: string, signal?: AbortSignal) =>
  request<Dungeon>(dungeonPath(name), { signal })
export const createDungeon = (name: string, monsters: number, difficulty: Difficulty) =>
  request<Dungeon>('/dungeons', {
    method: 'POST',
    body: JSON.stringify({ name, monsters, difficulty, heroClass: 'warrior' }),
  })
export const deleteDungeon = (name: string, signal?: AbortSignal) =>
  request<void>(dungeonPath(name), { method: 'DELETE', signal })
export const submitAttack = (dungeon: Dungeon, target: string, signal?: AbortSignal) =>
  request<Dungeon>(`${dungeonPath(dungeon.metadata.name)}/attacks`, {
    method: 'POST',
    body: JSON.stringify({ target, seq: dungeon.spec.attackSeq }),
    signal,
  })
export const openTreasure = (dungeon: Dungeon, signal?: AbortSignal) =>
  submitAttack(dungeon, 'open-treasure', signal)
export const getResources = (name: string, signal?: AbortSignal) =>
  request<KubernetesResource[]>(`${dungeonPath(name)}/resources`, { signal })
export const evaluateCEL = (name: string, expr: string, signal?: AbortSignal) =>
  request<{ result: string; error?: string }>(`${dungeonPath(name)}/cel-eval`, {
    method: 'POST',
    body: JSON.stringify({ expr }),
    signal,
  })

export function isAborted(error: unknown) {
  return error instanceof DOMException && error.name === 'AbortError'
}

export function errorMessage(error: unknown) {
  if (error instanceof DOMException && error.name === 'TimeoutError') {
    return 'The local API timed out. Check the backend and try again.'
  }
  if (error instanceof SyntaxError) {
    return 'Unexpected API response. Check that /api is routed to the local backend.'
  }
  return error instanceof Error ? error.message : String(error)
}
