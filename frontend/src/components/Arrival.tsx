import { useEffect, useState, type ReactNode } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { warmUp } from '../startup';

export function Arrival({ children }: { children: ReactNode }) {
  const client = useQueryClient();
  const [attempt, setAttempt] = useState(0);
  const [phase, setPhase] = useState<'loading' | 'error' | 'ready'>('loading');
  useEffect(() => {
    const controller = new AbortController();
    let active = true;
    const timeout = window.setTimeout(() => {
      if (active) setPhase('error');
      controller.abort();
    }, 10000);
    void warmUp(client, controller.signal)
      .then(() => {
        if (active && !controller.signal.aborted) setPhase('ready');
      })
      .catch(() => {
        if (active) setPhase('error');
      })
      .finally(() => window.clearTimeout(timeout));
    return () => {
      active = false;
      window.clearTimeout(timeout);
    };
  }, [client, attempt]);
  useEffect(() => {
    if (phase === 'ready') return;
    const previous = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.body.style.overflow = previous;
    };
  }, [phase]);
  return (
    <>
      <div
        className={`arrival-content ${phase === 'ready' ? 'has-arrived' : ''}`}
        inert={phase !== 'ready'}
      >
        {children}
      </div>
      {phase !== 'ready' && (
        <div className="arrival-screen" aria-label="Preparing your space">
          <div className="arrival-book" aria-hidden="true">
            <span />
            <span />
            <span />
          </div>
          <p className="arrival-brand">lector</p>
          <p role="status">
            {phase === 'error'
              ? 'Your space is taking a little longer.'
              : 'Settling into your space…'}
          </p>
          <p className="arrival-note">
            {phase === 'error'
              ? 'Check your connection, or continue while it loads.'
              : 'Preparing your library, pages, and type.'}
          </p>
          {phase === 'error' && (
            <div className="arrival-actions">
              <button
                className="button primary"
                onClick={() => {
                  setPhase('loading');
                  setAttempt((value) => value + 1);
                }}
              >
                Try again
              </button>
              <button className="button secondary" onClick={() => setPhase('ready')}>
                Continue
              </button>
            </div>
          )}
        </div>
      )}
    </>
  );
}
