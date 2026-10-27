import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, BookOpen, Check, Pencil, Plus, Trash2 } from 'lucide-react';
import { request, useSaveItem } from '../api';
import { percent, statusLabels, unit, type Item } from '../types';
import { Cover } from '../components/Cover';
import { ErrorMessage, Loading } from '../components/Feedback';
import { ItemForm } from '../components/ItemForm';
export default function Detail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const cache = useQueryClient();
  const [editing, setEditing] = useState(false);
  const save = useSaveItem();
  const query = useQuery({ queryKey: ['item', id], queryFn: () => request<Item>(`/items/${id}`) });
  const remove = useMutation({
    mutationFn: () => request(`/items/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      void cache.invalidateQueries();
      void navigate('/library');
    },
  });
  if (query.isPending) return <Loading />;
  if (query.error)
    return (
      <div className="page">
        <ErrorMessage error={query.error} />
        <Link to="/library">Back to library</Link>
      </div>
    );
  const item = query.data!;
  return (
    <div className="page detail-page">
      <Link to="/library" className="back-link">
        <ArrowLeft size={16} />
        Back to your library
      </Link>
      <div className="detail-layout">
        <div className="detail-cover">
          <Cover item={item} />
        </div>
        <div className="detail-copy">
          <span className="eyebrow">
            {item.kind} {item.format ? ` / ${item.format.toUpperCase()}` : ''}
          </span>
          <h1>{item.title}</h1>
          <p className="detail-author">{item.author || 'Unknown creator'}</p>
          <span className={`status-badge status-${item.status}`}>{statusLabels[item.status]}</span>
          <div className="detail-progress">
            <div>
              <h2>Your progress</h2>
              <strong>
                {item.current} / {item.total || '—'} {unit(item.kind)}
              </strong>
            </div>
            <progress
              max={item.total || Math.max(1, item.current)}
              value={item.current}
              aria-label="Reading progress"
            />
            <span>
              {item.total ? `${percent(item)}% complete` : 'Set a total to track your progress'}
            </span>
          </div>
          <div className="detail-actions">
            {item.file ? (
              <Link to={`/items/${item.id}/read`} className="button primary">
                <BookOpen size={18} />
                {item.current ? 'Continue reading' : 'Open reader'}
              </Link>
            ) : null}
            <button
              className="button secondary"
              disabled={save.isPending || (item.total > 0 && item.current >= item.total)}
              onClick={() =>
                save.mutate({
                  ...item,
                  current: item.current + 1,
                  status:
                    item.total > 0 && item.current + 1 >= item.total ? 'completed' : 'reading',
                })
              }
            >
              <Plus size={16} />
              Log a {item.kind === 'video' || item.kind === 'anime' ? 'episode' : 'chapter'}
            </button>
            <button
              className="button secondary"
              disabled={save.isPending || item.status === 'completed'}
              onClick={() =>
                save.mutate({ ...item, status: 'completed', current: item.total || item.current })
              }
            >
              <Check size={16} />
              Mark complete
            </button>
          </div>
          <ErrorMessage error={save.error} />
          <ErrorMessage error={remove.error} />
          <dl className="detail-meta">
            <div>
              <dt>Collection</dt>
              <dd>{item.collection || 'Not on a collection shelf'}</dd>
            </div>
            <div>
              <dt>Added</dt>
              <dd>
                {new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(
                  new Date(item.created_at),
                )}
              </dd>
            </div>
          </dl>
          <section className="notes-section">
            <h2>Your notes</h2>
            <p>
              {item.notes ||
                'A favorite line, a passing thought, something to come back to. Add a note when you edit this title.'}
            </p>
          </section>
          <div className="detail-bottom">
            <button className="text-button" onClick={() => setEditing(true)}>
              <Pencil size={15} />
              Edit title & notes
            </button>
            <button
              className="text-button danger"
              disabled={remove.isPending}
              onClick={() => {
                if (
                  window.confirm(
                    `Remove “${item.title}” and its imported file from your library? This cannot be undone.`,
                  )
                )
                  remove.mutate();
              }}
            >
              <Trash2 size={15} />
              Remove title
            </button>
          </div>
        </div>
      </div>
      {editing ? <ItemForm item={item} close={() => setEditing(false)} /> : null}
    </div>
  );
}
