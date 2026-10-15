import { useEffect, useRef, useState } from 'react'
import {
  ApiError, deleteDungeon, errorMessage, getDungeon, isAborted, openTreasure, submitAttack,
  type Dungeon,
} from './api'
import { DIFFICULTIES, dungeonOutcome, inspectProjection } from './game'
import { PixelIcon } from './PixelIcon'
import { ResourceInspector } from './ResourceInspector'
import { Sprite, type SpriteAction } from './Sprite'
import { CelPlayground, TeachingPanel } from './Teaching'

function HealthBar({ label, hp, max }: { label: string; hp: number; max: number }) {
  const percent = Math.max(0, Math.min(100, hp / max * 100))
  return (
    <div className="health">
      <div className="health-label"><span>{label}</span>
        <output aria-label={label} data-testid={label === 'Hero HP' ? 'hero-hp' : undefined}>{hp} / {max} HP</output>
      </div>
      <div className="health-track" aria-hidden="true">
        <div className={`health-fill ${percent < 25 ? 'low' : percent < 50 ? 'mid' : ''}`} style={{ width: `${percent}%` }} />
      </div>
    </div>
  )
}

export function DungeonView({ initialDungeon, onUpdated, onDeleted, onUnauthorized, onBack }: {
  initialDungeon: Dungeon
  onUpdated: (dungeon: Dungeon) => void
  onDeleted: (name: string) => void
  onUnauthorized: () => void
  onBack: () => void
}) {
  const [dungeon, setDungeon] = useState(initialDungeon)
  const [pollError, setPollError] = useState('')
  const [actionError, setActionError] = useState('')
  const [needsRefresh, setNeedsRefresh] = useState(false)
  const [notice, setNotice] = useState('')
  const [busyTarget, setBusyTarget] = useState<string | null>(null)
  const [revision, setRevision] = useState(0)
  const [flash, setFlash] = useState<{ target: string; heroHurt: boolean } | null>(null)
  const pending = useRef(false)
  const actionVersion = useRef(0)
  const alive = useRef(true)
  const actionController = useRef<AbortController | null>(null)
  const flashTimer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined)
  const name = initialDungeon.metadata.name

  useEffect(() => {
    alive.current = true
    return () => {
      alive.current = false
      actionController.current?.abort()
      clearTimeout(flashTimer.current)
    }
  }, [])

  useEffect(() => {
    let active = true
    let timer: ReturnType<typeof setTimeout>
    const controller = new AbortController()
    async function poll() {
      const version = actionVersion.current
      if (!pending.current) {
        try {
          const snapshot = await getDungeon(name, controller.signal)
          // A poll started before a mutation must not replace its newer response.
          if (active && !pending.current && version === actionVersion.current) {
            setDungeon(snapshot)
            onUpdated(snapshot)
            setPollError('')
            setNeedsRefresh(false)
          }
        } catch (failure) {
          if (active && !isAborted(failure) && version === actionVersion.current) {
            if (failure instanceof ApiError && failure.status === 401) onUnauthorized()
            else setPollError(errorMessage(failure))
          }
        }
      }
      if (active) timer = setTimeout(poll, 1500)
    }
    void poll()
    return () => { active = false; controller.abort(); clearTimeout(timer) }
  }, [name, revision, onUpdated, onUnauthorized])

  async function act(target: string) {
    if (pending.current) return
    pending.current = true
    actionVersion.current += 1
    const controller = new AbortController()
    actionController.current = controller
    setBusyTarget(target)
    setActionError('')
    setNotice('')
    try {
      if (target === 'delete') {
        await deleteDungeon(name, controller.signal)
        if (alive.current) onDeleted(name)
        return
      }
      const next = target === 'open-treasure'
        ? await openTreasure(dungeon, controller.signal)
        : await submitAttack(dungeon, target, controller.signal)
      if (!alive.current) return
      setDungeon(next)
      onUpdated(next)
      setNotice(target === 'open-treasure' ? 'Treasure opened in spec.game. Inspect the composed loot resources.'
        : `Turn ${next.spec.attackSeq} saved by Go in spec.game.`)
      clearTimeout(flashTimer.current)
      setFlash({ target, heroHurt: next.spec.game.heroHP < dungeon.spec.game.heroHP })
      flashTimer.current = setTimeout(() => setFlash(null), 650)
    } catch (failure) {
      if (!alive.current || isAborted(failure)) return
      if (failure instanceof ApiError && failure.status === 401) onUnauthorized()
      else if (failure instanceof ApiError && failure.status === 409) {
        setNeedsRefresh(true)
        setActionError(`Turn conflict (409): ${failure.message} Latest state is being reloaded.
          No action was automatically replayed; choose your next action again.`)
      } else setActionError(errorMessage(failure))
    } finally {
      pending.current = false
      actionController.current = null
      if (alive.current) { setBusyTarget(null); setRevision(value => value + 1) }
    }
  }

  const { game, difficulty, attackSeq } = dungeon.spec
  const rules = DIFFICULTIES[difficulty]
  const outcome = dungeonOutcome(dungeon)
  const terminal = outcome !== 'In progress'
  const { living, phase, issues } = inspectProjection(dungeon)
  const busy = busyTarget !== null
  const blocked = busy || Boolean(pollError) || needsRefresh
  const bossTarget = `${name}-boss`
  const heroAction: SpriteAction = game.heroHP <= 0 ? 'dead' : outcome === 'Victory' ? 'victory'
    : busyTarget && busyTarget !== 'open-treasure' && busyTarget !== 'delete' ? 'attack'
      : flash?.heroHurt ? 'hurt' : 'idle'
  const bossAction: SpriteAction = game.bossHP <= 0 ? 'dead' : flash?.target === bossTarget ? 'hurt' : 'idle'

  return (
    <>
      <div className="dungeon-heading">
        <div>
          <button className="text-button" onClick={onBack}>← All dungeons</button>
          <h2>{name}</h2>
          <p className="muted">rpg-demo · {difficulty} · Warrior · Sequence {attackSeq}</p>
        </div>
        <button className="button danger small" disabled={busy} onClick={() => void act('delete')}>
          {busyTarget === 'delete' ? 'Deleting…' : 'Delete dungeon'}
        </button>
      </div>
      {pollError && <p className="error" role="alert">Dungeon refresh failed: {pollError}
        {' '}Showing the last snapshot; attacks are paused until a refresh succeeds.</p>}
      {actionError && <p className="error" role="alert">{actionError}</p>}
      <div className="demo-layout">
        <div className="game-column">
          <section className="panel battle" aria-label="Dungeon arena">
            <div className="section-heading">
              <h2><PixelIcon name="sword" /> The dungeon</h2>
              <span className="badge">{living} / {game.monsterHP.length} monsters alive</span>
            </div>
            <p className="muted">Defeat the monsters, unlock the boss, then reveal the treasure.
              {' '}Go applies {rules.damage} attack damage on {difficulty}; HP below is read from <code>spec.game</code>.</p>
            <div className="arena">
              <div className="arena-entity">
                <Sprite spriteType="warrior" action={heroAction} size={112} />
                <h3>Warrior</h3>
                <HealthBar label="Hero HP" hp={game.heroHP} max={200} />
                <p className={game.heroHP <= 0 ? 'state dead-text' : 'state'}>{game.heroHP <= 0 ? 'Defeated' : 'Your hero'}</p>
              </div>
              <span className="arena-versus" aria-hidden="true">VS</span>
              <div className={`arena-entity ${living > 0 && game.bossHP > 0 ? 'locked' : ''}`}>
                <Sprite spriteType="dragon" action={bossAction} size={128} flip />
                <h3>Dragon boss</h3>
                <HealthBar label="Boss HP" hp={game.bossHP} max={rules.bossHP} />
                <p className="state" aria-label="Boss state">{game.bossHP <= 0 ? 'Defeated'
                  : living > 0 ? 'Locked — defeat all monsters' : terminal ? 'Battle ended' : 'Ready to fight'}</p>
                <button className="button primary" aria-label="Attack boss"
                  disabled={blocked || terminal || living > 0 || game.bossHP <= 0}
                  onClick={() => void act(bossTarget)}>Attack boss</button>
              </div>
            </div>
            <div className="monster-grid">
              {game.monsterHP.map((hp, index) => {
                const target = `${name}-monster-${index}`
                return (
                  <div className={`monster-card ${hp <= 0 ? 'defeated' : ''}`} key={target}>
                    <Sprite spriteType={index % 2 === 0 ? 'goblin' : 'skeleton'}
                      action={hp <= 0 ? 'dead' : flash?.target === target ? 'hurt' : 'idle'} size={80} />
                    <h3>Monster {index + 1}</h3>
                    <HealthBar label={`Monster ${index + 1} HP`} hp={hp} max={rules.monsterHP} />
                    <p className="state">{hp <= 0 ? 'Defeated' : terminal ? 'Battle ended' : 'Alive'}</p>
                    <button className="button primary small" aria-label={`Attack monster ${index + 1}`}
                      disabled={blocked || terminal || hp <= 0} onClick={() => void act(target)}>
                      Attack monster {index + 1}
                    </button>
                  </div>
                )
              })}
            </div>
            <div className="turn-notice" role="status">{busy && busyTarget !== 'delete' ? 'Saving action…' : notice}</div>
            {terminal && (
              <div className={`outcome ${outcome === 'Victory' ? 'win' : 'loss'}`} role="status" aria-label="Game outcome" data-testid="outcome">
                <h3><PixelIcon name={outcome === 'Victory' ? 'crown' : 'skull'} size={24} /> {outcome}</h3>
                <p>{outcome === 'Victory' ? 'The boss is defeated. This is the final room.' : 'Your warrior has fallen. Create another dungeon to try again.'}</p>
              </div>
            )}
            {outcome === 'Victory' && (
              <div className="treasure">
                <img src={`/sprites/dungeon/chest-${game.treasureOpened === 1 ? 'opened' : 'closed'}.png`}
                  alt={game.treasureOpened === 1 ? 'Opened treasure chest' : 'Closed treasure chest'} width={64} height={64} />
                {game.treasureOpened === 1
                  ? <p>Treasure opened. The Loot and Secret objects appear as kro reconciles.</p>
                  : <button className="button gold" disabled={blocked} onClick={() => void act('open-treasure')}>
                    Open treasure
                  </button>}
              </div>
            )}
          </section>
          <section className="panel projection" aria-labelledby="projection-title">
            <h2 id="projection-title">kro status projection</h2>
            <dl className="projection-values">
              <div><dt>Living monsters</dt><dd>{dungeon.status?.livingMonsters ?? 'Not reported'}</dd></div>
              <div><dt>Boss state</dt><dd>{dungeon.status?.bossState ?? 'Not reported'}</dd></div>
              <div><dt>Boss phase</dt><dd data-testid="boss-phase">{dungeon.status?.bossPhase ?? 'Not reported'}
                {game.bossHP <= 0 && ' (boss defeated)'}</dd></div>
              <div><dt>Treasure state</dt><dd>{dungeon.status?.treasureState ?? 'Not reported'}</dd></div>
            </dl>
            {issues.length > 0
              ? <div className="reconciling" role="status"><strong>Reconciling — status is behind or not ready.</strong>
                <p>Values above are the last observed kro projection, not current combat state.
                  Current <code>spec.game</code> implies boss phase <strong>{phase}</strong>.</p>
                <ul>{issues.map(issue => <li key={issue}>{issue}</li>)}</ul>
              </div>
              : <p className="muted">Compared status fields match <code>spec.game</code>.
                Child resources may still be reconciling; inspect their actual snapshots.</p>}
            {(dungeon.status?.conditions?.length ?? 0) > 0 && (
              <details className="conditions"><summary>Controller conditions</summary>
                <ul>{dungeon.status?.conditions?.map((condition, index) => (
                  <li key={`${condition.type}-${index}`}><strong>{condition.type}: {condition.status}</strong>
                    {condition.reason && ` (${condition.reason})`}{condition.message && ` — ${condition.message}`}</li>
                ))}</ul>
              </details>
            )}
          </section>
          <TeachingPanel />
          <CelPlayground name={name} sequence={attackSeq} onUnauthorized={onUnauthorized} />
        </div>
        <ResourceInspector name={name} revision={revision} onUnauthorized={onUnauthorized} />
      </div>
    </>
  )
}
