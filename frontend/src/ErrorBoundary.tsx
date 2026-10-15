import React from 'react'

interface State {
  error: Error | null
}

export class ErrorBoundary extends React.Component<
  { children: React.ReactNode },
  State
> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  render() {
    if (this.state.error) {
      return (
        <main className="app">
          <section className="panel" role="alert">
            <h1>The demo could not render.</h1>
            <p className="error">{this.state.error.message}</p>
            <button className="button gold" onClick={() => window.location.reload()}>Reload demo</button>
          </section>
        </main>
      )
    }
    return this.props.children
  }
}
