import { useDeferredValue } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowDownToLine,
  ArrowRight,
  BookOpen,
  Grid2X2,
  List,
  Plus,
  Search,
  SlidersHorizontal,
  Sparkles,
} from 'lucide-react';
import { request } from '../api';
import {
  kindLabels,
  percent,
  statusLabels,
  unit,
  type Kind,
  type Page,
  type Stats,
  type Status,
} from '../types';
import { Cover } from '../components/Cover';
import { ErrorMessage, Loading } from '../components/Feedback';
export default function Library({ onAdd, onImport }: { onAdd: () => void; onImport: () => void }) {
  const [params, setParams] = useSearchParams();
  const query = useDeferredValue(params.get('q') || '');
  const kind = params.get('kind') || '';
  const status = params.get('status') || '';
  const collection = params.get('collection') || '';
  const sort = params.get('sort') || 'updated';
  const offset = Number(params.get('offset')) || 0;
  const view = params.get('view') || 'grid';
  const filter = new URLSearchParams({
    q: query,
    kind,
    status,
    collection,
    sort,
    offset: String(offset),
    limit: '24',
  }).toString();
  const items = useQuery({
    queryKey: ['items', filter],
    queryFn: () => request<Page>(`/items?${filter}`),
  });
  const stats = useQuery({ queryKey: ['stats'], queryFn: () => request<Stats>('/stats') });
  const recent = useQuery({
    queryKey: ['items', 'continue'],
    queryFn: () => request<Page>('/items?status=reading&limit=3'),
  });
  function change(key: string, value: string) {
    setParams(
      (prev) => {
        const next = new URLSearchParams(prev);
        if (value) next.set(key, value);
        else next.delete(key);
        if (key !== 'offset') next.delete('offset');
        return next;
      },
      { replace: key === 'q' },
    );
  }
  const title = collection || (status ? statusLabels[status as Status] : 'Your library');
  const filtered = !!(query || kind || status || collection);
  return (
    <div className="page library-page">
      <header className="page-heading">
        <div>
          <p className="eyebrow">A LITTLE SPACE FOR YOUR STORIES</p>
          <h1>
            {title || 'Your library'}
            <span className="heading-dot">.</span>
          </h1>
          <p className="page-description">Good stories deserve a place to stay.</p>
        </div>
        <div className="heading-actions">
          <button className="button secondary" onClick={onImport}>
            <ArrowDownToLine size={16} />
            Import
          </button>
          <button className="button primary" onClick={onAdd}>
            <Plus size={17} />
            Add a title
          </button>
        </div>
      </header>
      {!filtered ? (
        <>
          <section className="welcome-card" aria-label="Your reading space">
            <div className="welcome-copy">
              <span className="small-tag">
                <span className="local-dot" />A LIBRARY THAT’S YOURS
              </span>
              <h2>
                A new chapter
                <br />
                starts here.
              </h2>
              <p>
                Books to get lost in. Worlds to come back to.
                <br />
                Keep everything you love in one quiet place.
              </p>
              <button className="text-button" onClick={onAdd}>
                Add your next story <ArrowRight size={16} />
              </button>
            </div>
            <div className="book-scene" aria-hidden="true">
              <span className="scene-circle" />
              <div className="scene-book book-one">
                <span>
                  THE ART
                  <br />
                  OF SLOW
                  <br />
                  LIVING
                </span>
                <i />
                <small>A LITTLE MORE PRESENT</small>
              </div>
              <div className="scene-book book-two">
                <span>
                  somewhere,
                  <br />
                  <em>elsewhere.</em>
                </span>
                <div className="book-sun" />
                <small>STORIES TO WANDER IN</small>
              </div>
              <div className="scene-book book-three">
                <span>
                  ONE
                  <br />
                  MORE
                  <br />
                  CHAPTER
                </span>
                <BookOpen size={34} strokeWidth={1} />
                <small>STAY A LITTLE LONGER</small>
              </div>
              <span className="scene-caption">A world of stories. A space of your own.</span>
            </div>
          </section>
          <div className="stats-row">
            <div>
              <span className="stat-icon">
                <BookOpen size={18} />
              </span>
              <strong>{stats.data?.total ?? '—'}</strong>
              <span>titles in your library</span>
            </div>
            <span className="stat-divider" />
            <div>
              <span className="stat-dot green" />
              <strong>{stats.data?.reading ?? '—'}</strong>
              <span>in progress</span>
            </div>
            <span className="stat-divider" />
            <div>
              <span className="stat-dot gold" />
              <strong>{stats.data?.completed ?? '—'}</strong>
              <span>stories completed</span>
            </div>
            <span className="stats-end">
              Every page counts <Sparkles size={14} />
            </span>
          </div>
          {recent.data?.items.length ? (
            <section className="continue-section">
              <div className="section-heading">
                <h2>Pick up where you left off</h2>
                <Link to="/library?status=reading">
                  View all <ArrowRight size={14} />
                </Link>
              </div>
              <div className="continue-grid">
                {recent.data.items.map((item) => (
                  <Link key={item.id} to={`/items/${item.id}`} className="continue-card">
                    <Cover item={item} compact />
                    <div>
                      <span className="item-kind">{item.kind}</span>
                      <h3>{item.title}</h3>
                      <p>{item.author || 'Your personal library'}</p>
                      <div className="progress-track">
                        <span style={{ width: `${percent(item)}%` }} />
                      </div>
                      <small>
                        {item.current} of {item.total || '—'} {unit(item.kind)}
                      </small>
                    </div>
                    <ArrowRight size={17} />
                  </Link>
                ))}
              </div>
            </section>
          ) : null}
        </>
      ) : null}
      <section className="shelf-section" aria-labelledby="shelf-heading">
        <div className="section-heading">
          <h2 id="shelf-heading">
            {filtered ? 'On this shelf' : 'All your stories'}{' '}
            <span className="count-pill">{items.data?.total ?? 0}</span>
          </h2>
          <div className="view-switch" aria-label="Library layout">
            <button
              aria-label="Grid view"
              aria-pressed={view === 'grid'}
              className={view === 'grid' ? 'selected' : ''}
              onClick={() => change('view', 'grid')}
            >
              <Grid2X2 size={16} />
            </button>
            <button
              aria-label="List view"
              aria-pressed={view === 'list'}
              className={view === 'list' ? 'selected' : ''}
              onClick={() => change('view', 'list')}
            >
              <List size={18} />
            </button>
          </div>
        </div>
        <div className="shelf-toolbar">
          <div className="format-tabs" aria-label="Filter by format">
            <button
              className={!kind ? 'selected' : ''}
              onClick={() => change('kind', '')}
              aria-pressed={!kind}
            >
              All formats
            </button>
            {Object.entries(kindLabels).map(([key, label]) => (
              <button
                key={key}
                className={kind === key ? 'selected' : ''}
                aria-pressed={kind === key}
                onClick={() => change('kind', key)}
              >
                {label}
              </button>
            ))}
          </div>
          <div className="shelf-tools">
            <label className="search-field">
              <Search size={16} aria-hidden="true" />
              <input
                aria-label="Search your library"
                name="search"
                type="search"
                placeholder="Search your library…"
                value={params.get('q') || ''}
                onChange={(e) => change('q', e.target.value)}
                autoComplete="off"
              />
            </label>
            <label className="sort-field">
              <SlidersHorizontal size={15} aria-hidden="true" />
              <select
                aria-label="Sort titles"
                value={sort}
                onChange={(e) => change('sort', e.target.value)}
              >
                <option value="updated">Recently updated</option>
                <option value="added">Recently added</option>
                <option value="title">Title A–Z</option>
              </select>
            </label>
          </div>
        </div>
        <ErrorMessage error={items.error} retry={() => void items.refetch()} />
        {items.isPending ? (
          <Loading />
        ) : items.data?.items.length ? (
          <div className={view === 'list' ? 'item-list' : 'item-grid'}>
            {items.data.items.map((item) => (
              <Link to={`/items/${item.id}`} className="item-card" key={item.id}>
                <div className="item-cover-wrap">
                  <Cover item={item} />
                  <span className={`status-badge status-${item.status}`}>
                    {statusLabels[item.status]}
                  </span>
                </div>
                <div className="item-card-info">
                  <span className="item-kind">{kindLabels[item.kind as Kind]}</span>
                  <h3>{item.title}</h3>
                  <p>{item.author || 'Unknown creator'}</p>
                  <div className="item-progress">
                    <div className="progress-track">
                      <span style={{ width: `${percent(item)}%` }} />
                    </div>
                    <span>
                      {item.total ? `${percent(item)}%` : `${item.current} ${unit(item.kind)}`}
                    </span>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        ) : !items.error ? (
          <div className="empty-state">
            <div className="empty-icon">
              <BookOpen size={30} strokeWidth={1.3} />
              <span>+</span>
            </div>
            <h3>{filtered ? 'No titles on this shelf yet' : 'Your next favorite belongs here'}</h3>
            <p>
              {filtered
                ? 'Try another search or add a title to this shelf.'
                : 'Start with a book you love, a series you’re following, or a file from your device.'}
            </p>
            <button className="button primary" onClick={onAdd}>
              <Plus size={16} />
              Add your first title
            </button>
            {filtered ? (
              <button className="text-button" onClick={() => setParams({})}>
                Clear filters
              </button>
            ) : (
              <button className="text-button" onClick={onImport}>
                Or import a local file <ArrowRight size={14} />
              </button>
            )}
          </div>
        ) : null}
        {(items.data?.total || 0) > 24 ? (
          <div className="pagination">
            <button
              className="button secondary"
              disabled={offset === 0}
              onClick={() => change('offset', String(Math.max(0, offset - 24)))}
            >
              Previous
            </button>
            <span>
              Page {Math.floor(offset / 24) + 1} of {Math.ceil((items.data?.total || 0) / 24)}
            </span>
            <button
              className="button secondary"
              disabled={offset + 24 >= (items.data?.total || 0)}
              onClick={() => change('offset', String(offset + 24))}
            >
              Next
            </button>
          </div>
        ) : null}
      </section>
      <footer className="page-footer">
        <span>Made for the way you read.</span>
        <span>
          <span className="local-dot" />
          No account. No noise. Just your library.
        </span>
      </footer>
    </div>
  );
}
