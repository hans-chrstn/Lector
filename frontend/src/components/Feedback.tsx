import { AlertCircle, RefreshCw } from 'lucide-react';
export function ErrorMessage({ error, retry }: { error: Error | null; retry?: () => void }) {
  if (!error) return null;
  return (
    <div className="error-message" role="alert">
      <AlertCircle size={18} aria-hidden="true" />
      <span>{error.message}</span>
      {retry ? (
        <button className="text-button" onClick={retry}>
          <RefreshCw size={14} aria-hidden="true" />
          Retry
        </button>
      ) : null}
    </div>
  );
}
export function Loading() {
  return (
    <div className="loading" role="status">
      <span className="loading-dot" />
      Opening your library…
    </div>
  );
}
