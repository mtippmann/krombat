import type { Difficulty, Dungeon } from './api'

export const DIFFICULTIES: Record<Difficulty, { monsterHP: number; bossHP: number; damage: number }> = {
  easy: { monsterHP: 30, bossHP: 200, damage: 25 },
  normal: { monsterHP: 50, bossHP: 400, damage: 35 },
  hard: { monsterHP: 80, bossHP: 800, damage: 45 },
}

export function dungeonOutcome(dungeon: Dungeon) {
  if (dungeon.spec.game.heroHP <= 0) return 'Defeat'
  if (dungeon.spec.game.bossHP <= 0) return 'Victory'
  return 'In progress'
}

export function inspectProjection(dungeon: Dungeon) {
  const { game, difficulty } = dungeon.spec
  const living = game.monsterHP.filter(hp => hp > 0).length
  const bossMax = DIFFICULTIES[difficulty].bossHP
  const phase = game.bossHP <= bossMax * 0.25 ? 'phase3'
    : game.bossHP <= bossMax * 0.5 ? 'phase2' : 'phase1'
  const expected = {
    livingMonsters: String(living),
    bossState: game.bossHP <= 0 ? 'defeated' : living > 0 ? 'pending' : 'ready',
    bossPhase: phase,
    victory: String(dungeonOutcome(dungeon) === 'Victory'),
  }
  const issues = Object.entries(expected).flatMap(([field, value]) => {
    const actual = dungeon.status?.[field as keyof typeof expected]
    return String(actual).toLowerCase() === value.toLowerCase()
      ? [] : [`${field}: ${actual ?? 'not reported'} → ${value}`]
  })
  if (game.treasureOpened === 1 && dungeon.status?.treasureState !== 'opened') {
    issues.push(`treasureState: ${dungeon.status?.treasureState ?? 'not reported'} → opened`)
  }
  const ready = dungeon.status?.conditions?.find(condition => condition.type === 'Ready')
  if (ready && ready.status !== 'True') issues.push(`Ready condition: ${ready.status}`)
  const observed = dungeon.status?.observedGeneration ?? ready?.observedGeneration
  if (observed !== undefined && observed < (dungeon.metadata.generation ?? 0)) {
    issues.push(`Observed generation ${observed} is behind generation ${dungeon.metadata.generation}.`)
  }
  return { living, phase, issues }
}
