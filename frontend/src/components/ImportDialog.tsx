import { useState, type FormEvent } from 'react';
import { Upload } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { request } from '../api';
import { Modal } from './Modal';
import { ErrorMessage } from './Feedback';
export function ImportDialog({ close }: { close: () => void }) {
  const cache = useQueryClient();
  const [name, setName] = useState('Choose a file from your device');
  const upload = useMutation({
    mutationFn: (data: FormData) => request('/import', { method: 'POST', body: data }),
    onSuccess: () => {
      void cache.invalidateQueries({ queryKey: ['items'] });
      void cache.invalidateQueries({ queryKey: ['stats'] });
      close();
    },
  });
  function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    upload.mutate(new FormData(e.currentTarget));
  }
  function dismiss() {
    if (!upload.isPending) close();
  }
  return (
    <Modal title="Bring your library home" close={dismiss}>
      <form className="item-form" onSubmit={submit}>
        <p className="muted">Import a file you own. It stays on your Lector server.</p>
        <label className="file-drop">
          <Upload size={32} aria-hidden="true" />
          <strong>{name}</strong>
          <span>EPUB, PDF, TXT, CBZ, MP4, or WebM</span>
          <input
            type="file"
            name="file"
            required
            accept=".epub,.pdf,.txt,.cbz,.mp4,.webm"
            onChange={(e) => setName(e.target.files?.[0]?.name || 'Choose a file from your device')}
          />
        </label>
        <ErrorMessage error={upload.error} />
        <div className="form-actions">
          <button
            type="button"
            className="button secondary"
            onClick={dismiss}
            disabled={upload.isPending}
          >
            Cancel
          </button>
          <button className="button primary" disabled={upload.isPending}>
            {upload.isPending ? 'Importing…' : 'Import file'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
