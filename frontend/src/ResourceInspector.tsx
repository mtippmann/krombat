import { useEffect, useState } from 'react'
import { ApiError, errorMessage, getResources, isAborted, type KubernetesResource } from './api'
import { PixelIcon } from './PixelIcon'

function resourceKey(resource: KubernetesResource) {
  return `${resource.apiVersion}/${resource.kind}/${resource.metadata.namespace}/${resource.metadata.name}`
}

export function ResourceInspector({ name, revision, onUnauthorized }: {
  name: string
  revision: number
  onUnauthorized: () => void
}) {
  const [resources, setResources] = useState<KubernetesResource[]>([])
  const [selectedKey, setSelectedKey] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)
  const [updatedAt, setUpdatedAt] = useState<Date | null>(null)
  const [refresh, setRefresh] = useState(0)

  useEffect(() => {
    let active = true
    let timer: ReturnType<typeof setTimeout>
    const controller = new AbortController()
    setLoading(true)
    async function poll() {
      try {
        const snapshot = await getResources(name, controller.signal)
        if (active) {
          setResources(snapshot)
          setUpdatedAt(new Date())
          setError('')
        }
      } catch (failure) {
        if (active && !isAborted(failure)) {
          if (failure instanceof ApiError && failure.status === 401) onUnauthorized()
          else setError(errorMessage(failure))
        }
      } finally {
        if (active) {
          setLoading(false)
          timer = setTimeout(poll, 2000)
        }
      }
    }
    void poll()
    return () => { active = false; controller.abort(); clearTimeout(timer) }
  }, [name, revision, refresh, onUnauthorized])

  const selected = resources.find(resource => resourceKey(resource) === selectedKey) ?? resources[0]

  return (
    <aside className="panel resource-inspector" aria-labelledby="inspector-title">
      <div className="section-heading">
        <h2 id="inspector-title"><PixelIcon name="scroll" /> Resource inspector</h2>
        <button className="button subtle small" disabled={loading} onClick={() => setRefresh(value => value + 1)}>
          Refresh resources
        </button>
      </div>
      <p className="muted">Actual Kubernetes objects returned by the local API, polled every 2 seconds.
        Only this Dungeon and its owned demo resources are included.</p>
      <p className="muted">Upstream kro tracks membership with ApplySet inventory and kro finalizers,
        not default ownerReferences. The <code>kro.run/instance-id</code> label identifies the immediate parent UID.</p>
      {error && <p role="alert" className="error">Resource refresh failed: {error}
        {updatedAt && ' Showing the last successful snapshot.'}</p>}
      {!updatedAt && loading && <p role="status">Loading resources…</p>}
      {updatedAt && <p className="snapshot-time">Snapshot {updatedAt.toLocaleTimeString()} · {resources.length} objects</p>}
      {updatedAt && resources.length === 0 && <p>No resources were returned. Waiting for reconciliation.</p>}
      <ul className="resource-list" aria-label="Kubernetes resources" data-testid="resource-list">
        {resources.map(resource => (
          <li key={resourceKey(resource)}>
            <button className="resource-choice" aria-pressed={selected === resource}
              aria-label={`Inspect ${resource.kind ?? 'resource'} ${resource.metadata.name}`}
              onClick={() => setSelectedKey(resourceKey(resource))}>
              <span className="resource-kind">{resource.kind ?? 'Unknown kind'}</span>
              <span>{resource.metadata.name}</span>
            </button>
          </li>
        ))}
      </ul>
      {selected && (
        <div className="resource-detail">
          <h3>{selected.kind} / {selected.metadata.name}</h3>
          <p className="muted">Raw JSON · no generated graph nodes or simulated state</p>
          {selected.kind === 'Secret' && <p className="secret-note">Demo loot only. Secret data is base64-encoded,
            not encrypted by that encoding.</p>}
          <pre className="resource-json" tabIndex={0} aria-label="Resource JSON">{JSON.stringify(selected, null, 2)}</pre>
        </div>
      )}
    </aside>
  )
}
