import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react'
import {
  ApiError, createDungeon, errorMessage, getMe, isAborted, listDungeons, login, logout,
  type AuthUser, type Difficulty, type Dungeon,
} from './api'
import { DungeonView } from './DungeonView'
import { dungeonOutcome } from './game'
import { PixelIcon } from './PixelIcon'
import { Sprite } from './Sprite'
import { TeachingPanel } from './Teaching'
import './App.css'

function suggestedName() {
  return `demo-${Date.now().toString(36)}`
}

export default function App() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [authLoading, setAuthLoading] = useState(true)
  const [authRevision, setAuthRevision] = useState(0)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState<string | null>(null)
  const pending = useRef(false)
  const [dungeons, setDungeons] = useState<Dungeon[]>([])
  const [selected, setSelected] = useState<Dungeon | null>(null)
  const [listLoading, setListLoading] = useState(true)
  const [listError, setListError] = useState('')
  const [listRevision, setListRevision] = useState(0)
  const [name, setName] = useState(suggestedName)
  const [monsters, setMonsters] = useState('1')
  const [difficulty, setDifficulty] = useState<Difficulty>('easy')

  const sessionExpired = useCallback(() => {
    setUser(null)
    setDungeons([])
    setSelected(null)
    setError('Your local session expired. Enter the demo again to continue.')
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    setAuthLoading(true)
    setError('')
    getMe(controller.signal).then(current => {
      if (!controller.signal.aborted) setUser(current)
    }).catch(failure => {
      if (!controller.signal.aborted && !isAborted(failure)) setError(errorMessage(failure))
    }).finally(() => {
      if (!controller.signal.aborted) setAuthLoading(false)
    })
    return () => controller.abort()
  }, [authRevision])

  useEffect(() => {
    if (!user) return
    let active = true
    let timer: ReturnType<typeof setTimeout>
    const controller = new AbortController()
    setListLoading(true)
    async function poll() {
      try {
        const snapshot = await listDungeons(controller.signal)
        if (active) { setDungeons(snapshot); setListError('') }
      } catch (failure) {
        if (active && !isAborted(failure)) {
          if (failure instanceof ApiError && failure.status === 401) sessionExpired()
          else setListError(errorMessage(failure))
        }
      } finally {
        if (active) { setListLoading(false); timer = setTimeout(poll, 3000) }
      }
    }
    void poll()
    return () => { active = false; controller.abort(); clearTimeout(timer) }
  }, [user, listRevision, sessionExpired])

  const updateDungeon = useCallback((dungeon: Dungeon) => {
    setDungeons(current => {
      const exists = current.some(item => item.metadata.name === dungeon.metadata.name)
      return exists ? current.map(item => item.metadata.name === dungeon.metadata.name ? dungeon : item)
        : [dungeon, ...current]
    })
  }, [])

  const removeDungeon = useCallback((dungeonName: string) => {
    setSelected(null)
    setDungeons(current => current.filter(item => item.metadata.name !== dungeonName))
    setListRevision(value => value + 1)
  }, [])

  async function enterDemo() {
    if (pending.current) return
    pending.current = true
    setBusy('login')
    setError('')
    try { setUser(await login()) }
    catch (failure) { setError(errorMessage(failure)) }
    finally { pending.current = false; setBusy(null) }
  }

  async function leaveDemo() {
    if (pending.current) return
    pending.current = true
    setBusy('logout')
    setError('')
    try {
      await logout()
      setUser(null)
      setSelected(null)
      setDungeons([])
    } catch (failure) { setError(errorMessage(failure)) }
    finally { pending.current = false; setBusy(null) }
  }

  async function create(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (pending.current) return
    const count = Number(monsters)
    if (!Number.isInteger(count) || count < 1 || count > 10) {
      setError('Choose a whole number of monsters between 1 and 10.')
      return
    }
    pending.current = true
    setBusy('create')
    setError('')
    try {
      const dungeon = await createDungeon(name.trim(), count, difficulty)
      updateDungeon(dungeon)
      setSelected(dungeon)
      setName(suggestedName())
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) sessionExpired()
      else setError(errorMessage(failure))
    } finally { pending.current = false; setBusy(null) }
  }

  return (
    <div className="app">
      <a className="skip-link" href="#main">Skip to content</a>
      <header className="site-header">
        <div className="brand">
          <span aria-hidden="true"><PixelIcon name="shield" size={36} /></span>
          <div><h1>Krombat</h1><p>Local demo · vanilla kro</p></div>
        </div>
        <div className="session">
          <span className="badge">rpg-demo</span>
          {user && <><span>{user.login}</span>
            <button className="button subtle small" disabled={busy !== null} onClick={() => void leaveDemo()}>Log out</button></>}
        </div>
      </header>
      <main id="main">
        {error && <div className="error" role="alert">{error}
          {!user && !authLoading && <button className="text-button" onClick={() => setAuthRevision(value => value + 1)}>
            Retry connection
          </button>}
        </div>}
        {authLoading ? <p className="loading" role="status">Connecting to the local demo…</p>
          : !user ? (
            <>
              <section className="panel welcome">
                <div className="welcome-art" aria-hidden="true">
                  <Sprite spriteType="warrior" action="idle" size={128} />
                  <Sprite spriteType="dragon" action="idle" size={160} flip />
                </div>
                <p className="eyebrow">A dungeon made of Kubernetes resources</p>
                <h2>Play a turn. Watch a graph reconcile.</h2>
                <p>One warrior, one room, and real composed resources. Go handles combat;
                  upstream kro turns the desired game state into a live resource graph.</p>
                <button className="button gold" disabled={busy !== null} onClick={() => void enterDemo()}>
                  {busy === 'login' ? 'Entering…' : 'Enter local demo'}
                </button>
                <p className="muted">No GitHub account or external service. Entry creates a signed, HttpOnly
                  local session cookie; it is a demo identity, not production authentication.</p>
              </section>
              <TeachingPanel />
            </>
          ) : selected ? (
            <DungeonView key={selected.metadata.name} initialDungeon={selected} onUpdated={updateDungeon}
              onDeleted={removeDungeon} onUnauthorized={sessionExpired} onBack={() => setSelected(null)} />
          ) : (
            <>
              <section className="lobby-intro">
                <p className="eyebrow">Your local playground</p>
                <h2>Choose your dungeon</h2>
                <p className="muted">Start with one monster on easy for a quick, winnable tour.
                  All dungeons live in <code>rpg-demo</code>.</p>
              </section>
              <div className="lobby-layout">
                <section className="panel" aria-labelledby="create-title">
                  <h2 id="create-title"><PixelIcon name="sword" /> New dungeon</h2>
                  <p className="muted form-note">Warrior · 200 HP · One room</p>
                  <form className="create-form" onSubmit={create}>
                    <label htmlFor="dungeon-name">Dungeon name</label>
                    <input id="dungeon-name" value={name} required minLength={1} maxLength={40}
                      pattern={'[a-z0-9]([a-z0-9\\-]*[a-z0-9])?'} autoComplete="off" spellCheck={false}
                      aria-describedby="name-help" onChange={event => setName(event.target.value)} disabled={busy !== null} />
                    <small id="name-help" className="muted">Lowercase letters, numbers, and hyphens; start and end with a letter or number.</small>
                    <label htmlFor="monster-count">Monsters</label>
                    <input id="monster-count" type="number" min={1} max={10} step={1} value={monsters}
                      required onChange={event => setMonsters(event.target.value)} disabled={busy !== null} />
                    <label htmlFor="difficulty">Difficulty</label>
                    <select id="difficulty" value={difficulty} disabled={busy !== null}
                      onChange={event => setDifficulty(event.target.value as Difficulty)}>
                      <option value="easy">Easy</option><option value="normal">Normal</option><option value="hard">Hard</option>
                    </select>
                    <button className="button gold" type="submit" disabled={busy !== null}>
                      {busy === 'create' ? 'Creating…' : 'Create dungeon'}
                    </button>
                  </form>
                </section>
                <section className="panel" aria-labelledby="dungeons-title">
                  <div className="section-heading">
                    <h2 id="dungeons-title">Dungeons</h2>
                    <button className="button subtle small" disabled={listLoading}
                      onClick={() => setListRevision(value => value + 1)}>Refresh dungeons</button>
                  </div>
                  {listError && <p className="error" role="alert">Dungeon list refresh failed: {listError}</p>}
                  {listLoading && dungeons.length === 0 && <p role="status">Loading dungeons…</p>}
                  {!listLoading && !listError && dungeons.length === 0 && <p className="empty-state">No dungeons yet. Create your first adventure.</p>}
                  <ul className="dungeon-list" aria-label="Dungeons">
                    {dungeons.map(dungeon => (
                      <li key={dungeon.metadata.name}>
                        <button className="dungeon-choice" aria-label={dungeon.metadata.name}
                          onClick={() => { setError(''); setSelected(dungeon) }}>
                          <span className="dungeon-choice-title">{dungeon.metadata.name}</span>
                          <span className="muted">{dungeon.spec.difficulty} · {dungeon.spec.monsters} monsters
                            {' '}· Hero {dungeon.spec.game.heroHP} HP</span>
                          <span className={`outcome-label ${dungeonOutcome(dungeon).toLowerCase().replace(' ', '-')}`}>{dungeonOutcome(dungeon)}</span>
                        </button>
                      </li>
                    ))}
                  </ul>
                </section>
              </div>
              <TeachingPanel />
            </>
          )}
      </main>
      <footer className="site-footer">Go computes turns. Vanilla kro composes and reconciles. Your browser inspects.</footer>
    </div>
  )
}
