import { useEffect, useState, type FormEvent } from 'react';
import { useSaveItem } from '../api';
import { kindLabels, statusLabels, type Item, type Kind, type Status } from '../types';
import { ErrorMessage } from './Feedback';
import { Modal } from './Modal';
export function ItemForm({ item, close }: { item?: Item; close: () => void }) {
  const save = useSaveItem();
  const [dirty, setDirty] = useState(false);
  useEffect(() => {
    if (!dirty) return;
    function warn(event: BeforeUnloadEvent) {
      event.preventDefault();
    }
    window.addEventListener('beforeunload', warn);
    return () => window.removeEventListener('beforeunload', warn);
  }, [dirty]);
  function dismiss() {
    if (save.isPending) return;
    if (!dirty || window.confirm('Discard your unsaved changes?')) close();
  }
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const data = new FormData(e.currentTarget);
    try {
      await save.mutateAsync({
        ...item,
        title: String(data.get('title')),
        author: String(data.get('author')),
        kind: data.get('kind') as Kind,
        status: data.get('status') as Status,
        current: Number(data.get('current')),
        total: Number(data.get('total')),
        collection: String(data.get('collection')),
        notes: String(data.get('notes')),
      });
      close();
    } catch {
      return;
    }
  }
  return (
    <Modal title={item ? 'Edit your title' : 'Make room for a new story'} close={dismiss}>
      <form onSubmit={submit} onChange={() => setDirty(true)} className="item-form">
        <p className="muted">Track a book, series, or video in your own library.</p>
        <label>
          Title
          <input
            name="title"
            required
            maxLength={300}
            defaultValue={item?.title}
            autoComplete="off"
            placeholder="The title you’re thinking of…"
          />
        </label>
        <label>
          Author or creator
          <input
            name="author"
            maxLength={200}
            defaultValue={item?.author}
            autoComplete="off"
            placeholder="Who made it…"
          />
        </label>
        <div className="form-row">
          <label>
            Format
            <select name="kind" defaultValue={item?.kind || 'book'}>
              {Object.entries(kindLabels).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
            </select>
          </label>
          <label>
            Shelf
            <select name="status" defaultValue={item?.status || 'planned'}>
              {Object.entries(statusLabels).map(([key, label]) => (
                <option key={key} value={key}>
                  {label}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="form-row">
          <label>
            Current chapter / episode
            <input
              type="number"
              name="current"
              min={0}
              max={10000000}
              defaultValue={item?.current || 0}
              required
            />
          </label>
          <label>
            Total (0 if unknown)
            <input
              type="number"
              name="total"
              min={0}
              max={10000000}
              defaultValue={item?.total || 0}
              required
            />
          </label>
        </div>
        <label>
          Collection
          <input
            name="collection"
            maxLength={80}
            defaultValue={item?.collection}
            autoComplete="off"
            placeholder="A shelf of your own…"
          />
        </label>
        <label>
          Notes
          <textarea
            name="notes"
            maxLength={20000}
            rows={3}
            defaultValue={item?.notes}
            placeholder="Something to remember…"
          />
        </label>
        <ErrorMessage error={save.error} />
        <div className="form-actions">
          <button
            type="button"
            className="button secondary"
            onClick={dismiss}
            disabled={save.isPending}
          >
            Cancel
          </button>
          <button className="button primary" disabled={save.isPending}>
            {save.isPending ? 'Saving…' : item ? 'Save changes' : 'Add to library'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
