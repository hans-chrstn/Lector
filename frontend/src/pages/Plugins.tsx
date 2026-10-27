import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Check, Globe2, LayoutPanelTop, Play, Plus, Puzzle, ShieldCheck } from 'lucide-react';
import { request, useSaveItem } from '../api';
import type { Plugin, PluginResult } from '../types';
import { ErrorMessage, Loading } from '../components/Feedback';
import { Modal } from '../components/Modal';
function PluginCard({ plugin }: { plugin: Plugin }) {
  const cache = useQueryClient();
  const [review, setReview] = useState(false);
  const [input, setInput] = useState('');
  const [added, setAdded] = useState<number[]>([]);
  const save = useSaveItem();
  const toggle = useMutation({
    mutationFn: (enabled: boolean) =>
      request(`/plugins/${plugin.id}`, {
        method: 'PUT',
        body: JSON.stringify({ enabled, generation: plugin.generation }),
      }),
    onSuccess: () => {
      setReview(false);
      void cache.invalidateQueries({ queryKey: ['plugins'] });
    },
  });
  const run = useMutation({
    mutationFn: (action: string) =>
      request<PluginResult>(`/plugins/${plugin.id}/actions/${action}`, {
        method: 'POST',
        body: JSON.stringify({ input }),
      }),
    onMutate: () => setAdded([]),
  });
  async function add(index: number) {
    const entry = run.data!.entries[index];
    try {
      await save.mutateAsync({
        title: entry.title,
        author: entry.author,
        kind: entry.kind,
        total: entry.total,
        current: 0,
        status: 'planned',
        collection: '',
        notes: entry.text?.slice(0, 20000) || '',
      });
      setAdded((prev) => [...prev, index]);
    } catch {
      return;
    }
  }
  return (
    <article className="plugin-card">
      <div className="plugin-heading">
        <span className="plugin-icon">
          <Puzzle size={23} />
        </span>
        <div>
          <h2>{plugin.name}</h2>
          <span>Version {plugin.version || '—'}</span>
        </div>
        <span className={`plugin-state ${plugin.enabled ? 'enabled' : ''}`}>
          {plugin.enabled ? 'Enabled' : 'Disabled'}
        </span>
      </div>
      <p>{plugin.description || 'A personal extension for your library.'}</p>
      <div className="permission-summary">
        <span>
          <Globe2 size={14} />
          {plugin.hosts?.length || 0} network domains
        </span>
        <span>
          <LayoutPanelTop size={14} />
          {plugin.ui ? 'UI panels' : 'No UI access'}
        </span>
      </div>
      {plugin.error ? <ErrorMessage error={new Error(plugin.error)} /> : null}
      <ErrorMessage error={toggle.error} />
      <button
        className={`button ${plugin.enabled ? 'secondary' : 'primary'}`}
        disabled={toggle.isPending || !!plugin.error}
        onClick={() => {
          if (plugin.enabled) toggle.mutate(false);
          else setReview(true);
        }}
      >
        {toggle.isPending ? 'Saving…' : plugin.enabled ? 'Disable extension' : 'Review & enable'}
      </button>
      {plugin.enabled ? (
        <div className="plugin-workspace">
          <label>
            Input
            <input
              value={input}
              maxLength={16000}
              onChange={(e) => setInput(e.target.value)}
              placeholder="Enter a query or value…"
              autoComplete="off"
            />
          </label>
          <div className="detail-actions">
            {plugin.actions?.map((action) => (
              <button
                key={action.id}
                className="button secondary"
                disabled={run.isPending}
                onClick={() => run.mutate(action.id)}
              >
                <Play size={14} />
                {run.isPending ? 'Running…' : action.label}
              </button>
            ))}
          </div>
          <ErrorMessage error={run.error} />
          <ErrorMessage error={save.error} />
          <div aria-live="polite">
            {run.data?.panels.map((panel, i) => (
              <section
                key={i}
                className={`extension-panel ${panel.tone === 'accent' ? 'accent' : ''}`}
              >
                <h3>{panel.title}</h3>
                <p>{panel.text}</p>
              </section>
            ))}
            {run.data?.entries.map((entry, i) => (
              <div className="extension-entry" key={i}>
                <div>
                  <strong>{entry.title}</strong>
                  <small>
                    {entry.author} · {entry.kind}
                  </small>
                </div>
                <button
                  className="button secondary"
                  disabled={save.isPending || added.includes(i)}
                  onClick={() => void add(i)}
                >
                  {added.includes(i) ? <Check size={15} /> : <Plus size={15} />}{' '}
                  {added.includes(i) ? 'Added' : 'Add to library'}
                </button>
              </div>
            ))}
          </div>
        </div>
      ) : null}
      {review ? (
        <Modal title={`Enable ${plugin.name}?`} close={() => setReview(false)}>
          <div className="item-form">
            <p>This extension requests the following access. Enable only extensions you trust.</p>
            <ul className="permission-list">
              <li>
                <LayoutPanelTop size={18} />
                {plugin.ui ? 'Display text panels and action results' : 'No UI panels'}
              </li>
              {plugin.hosts?.map((host) => (
                <li key={host}>
                  <Globe2 size={18} />
                  HTTPS requests to <strong>{host}</strong>
                </li>
              ))}
              {!plugin.hosts?.length ? (
                <li>
                  <ShieldCheck size={18} />
                  No network access
                </li>
              ) : null}
            </ul>
            <p className="muted">
              Extensions cannot run browser scripts or read your library files. Permission changes
              disable the extension until you approve them again.
            </p>
            <ErrorMessage error={toggle.error} />
            <div className="form-actions">
              <button className="button secondary" onClick={() => setReview(false)}>
                Cancel
              </button>
              <button
                className="button primary"
                disabled={toggle.isPending}
                onClick={() => toggle.mutate(true)}
              >
                Approve & enable
              </button>
            </div>
          </div>
        </Modal>
      ) : null}
    </article>
  );
}
export default function Plugins() {
  const query = useQuery({
    queryKey: ['plugins'],
    queryFn: () => request<Plugin[]>('/plugins'),
    refetchInterval: 3000,
  });
  return (
    <div className="page">
      <header className="page-heading">
        <div>
          <p className="eyebrow">MAKE IT YOUR OWN</p>
          <h1>
            A little more possibility<span className="heading-dot">.</span>
          </h1>
          <p className="page-description">Personal extensions, with permission at the center.</p>
        </div>
      </header>
      <div className="info-banner">
        <ShieldCheck size={23} />
        <div>
          <strong>Your library sets the boundaries.</strong>
          <p>Extensions start disabled. You choose their network and display permissions.</p>
        </div>
      </div>
      <ErrorMessage error={query.error} retry={() => void query.refetch()} />
      {query.isPending ? (
        <Loading />
      ) : query.data?.length ? (
        <div className="plugin-grid">
          {query.data.map((plugin) => (
            <PluginCard
              key={`${plugin.id}-${plugin.generation}-${plugin.enabled}`}
              plugin={plugin}
            />
          ))}
        </div>
      ) : (
        <div className="empty-state">
          <div className="empty-icon">
            <Puzzle size={30} />
          </div>
          <h2>Room to grow</h2>
          <p>
            Install a personal extension in your local plugins folder. It will appear here
            automatically, ready for you to review.
          </p>
          <p className="muted">See API.md in your Lector project for the extension format.</p>
        </div>
      )}
    </div>
  );
}
