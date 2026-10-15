import { useEffect, useRef, useState, type FormEvent } from 'react'
import { ApiError, errorMessage, evaluateCEL, isAborted } from './api'
import { PixelIcon } from './PixelIcon'

export function TeachingPanel() {
  return (
    <section className="panel teaching" aria-labelledby="teaching-title">
      <h2 id="teaching-title"><PixelIcon name="book" /> What is kro doing?</h2>
      <p className="muted">A small game, a real Kubernetes resource graph, and no custom kro extensions.</p>
      <ol className="lesson-list">
        <li><strong>Go owns the turns.</strong> The API validates an attack, computes combat, and updates
          <code> spec.game</code> with an incremented <code>attackSeq</code>. CEL does not run the combat engine.</li>
        <li><strong>Composition &amp; <code>forEach</code>.</strong> Vanilla kro projects the Dungeon into Hero,
          Monster, Boss, Treasure, and Loot resources. A collection expands the monster array into child resources;
          child graphs compose ConfigMaps and demo Secrets.</li>
        <li><strong><code>includeWhen</code>.</strong> Optional treasure and loot resources appear only when
          their graph conditions allow them. The inspector never invents missing resources.</li>
        <li><strong>Status aggregation.</strong> CEL expressions summarize child outputs in Dungeon
          <code> status</code>. That projection can lag behind <code>spec.game</code> during reconciliation.</li>
        <li><strong>Drift repair.</strong> If a composed resource drifts from its desired template, kro
          reconciles it back. ApplySet inventory and kro finalizers manage composed resources; inspect
          <code> kro.run/instance-id</code> labels to find their immediate parent, rather than assuming ownerReference GC.</li>
      </ol>
    </section>
  )
}

const EXAMPLES = [
  ['Living monsters', 'schema.spec.game.monsterHP.filter(hp, hp > 0).size()'],
  ['Boss defeated', 'schema.spec.game.bossHP <= 0'],
  ['Hero HP', 'schema.spec.game.heroHP'],
  ['Treasure opened', 'schema.spec.game.treasureOpened == 1'],
] as const

export function CelPlayground({ name, sequence, onUnauthorized }: {
  name: string
  sequence: number
  onUnauthorized: () => void
}) {
  const [expression, setExpression] = useState<string>(EXAMPLES[0][1])
  const [evaluation, setEvaluation] = useState<{ result: string; sequence: number; error?: string } | null>(null)
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const controller = useRef<AbortController | null>(null)

  useEffect(() => () => controller.current?.abort(), [])

  async function evaluate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (controller.current || !expression.trim()) return
    const requestController = new AbortController()
    controller.current = requestController
    setBusy(true)
    setError('')
    setEvaluation(null)
    try {
      const response = await evaluateCEL(name, expression, requestController.signal)
      if (!requestController.signal.aborted) setEvaluation({ ...response, sequence })
    } catch (failure) {
      if (failure instanceof ApiError && failure.status === 401) onUnauthorized()
      else if (!isAborted(failure)) setError(errorMessage(failure))
    } finally {
      controller.current = null
      if (!requestController.signal.aborted) setBusy(false)
    }
  }

  return (
    <section className="panel cel-playground" aria-labelledby="cel-title">
      <h2 id="cel-title"><PixelIcon name="scroll" /> CEL playground</h2>
      <p className="muted">Evaluate read-only expressions against the latest Dungeon on the backend.
        <code> schema.spec.game</code> is the saved game; evaluating does not change it.</p>
      <div className="example-buttons" aria-label="CEL examples">
        {EXAMPLES.map(([label, value]) => (
          <button type="button" className="button subtle small" key={label} disabled={busy}
            onClick={() => { setExpression(value); setEvaluation(null); setError('') }}>{label}</button>
        ))}
      </div>
      <form onSubmit={evaluate}>
        <label htmlFor="cel-expression">CEL expression</label>
        <textarea id="cel-expression" value={expression} rows={3} required maxLength={2000}
          spellCheck={false} disabled={busy}
          onChange={event => { setExpression(event.target.value); setEvaluation(null); setError('') }} />
        <button className="button secondary" type="submit" disabled={busy || !expression.trim()}>
          {busy ? 'Evaluating…' : 'Evaluate CEL'}
        </button>
      </form>
      {error && <p role="alert" className="error">{error}</p>}
      {evaluation && (
        <div className={evaluation.error ? 'error' : 'cel-result'} role={evaluation.error ? 'alert' : 'status'}>
          {evaluation.error ? <p>{evaluation.error}</p> : <pre aria-label="CEL result" data-testid="cel-result">{evaluation.result}</pre>}
          <p className="muted">Evaluated when the UI last observed sequence {evaluation.sequence}.
            {evaluation.sequence !== sequence && ' A turn has changed since then; evaluate again.'}
            {' '}This result is a snapshot, not a live binding.</p>
        </div>
      )}
    </section>
  )
}
