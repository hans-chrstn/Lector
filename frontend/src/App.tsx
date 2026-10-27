import { lazy, Suspense, useState } from 'react';
import { NavLink, Route, Routes, Link, useLocation, Navigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  ArrowDownToLine,
  BookOpen,
  Bookmark,
  CheckCheck,
  Clock3,
  Folder,
  LibraryBig,
  Menu,
  Plus,
  Puzzle,
  Settings2,
  X,
} from 'lucide-react';
import { request } from './api';
import type { Stats } from './types';
import { Loading } from './components/Feedback';
import { ItemForm } from './components/ItemForm';
import { ImportDialog } from './components/ImportDialog';
import Library from './pages/Library';
import Landing from './pages/Landing';
import { routeModules } from './startup';
const Detail = lazy(routeModules.Detail);
const Reader = lazy(routeModules.Reader);
const Plugins = lazy(routeModules.Plugins);
const History = lazy(routeModules.History);
const Settings = lazy(routeModules.Settings);
export default function App() {
  const [dialog, setDialog] = useState<'add' | 'import' | null>(null);
  const [menu, setMenu] = useState(false);
  const location = useLocation();
  const stats = useQuery({ queryKey: ['stats'], queryFn: () => request<Stats>('/stats') });
  const shelf = new URLSearchParams(location.search).get('status');
  const collection = new URLSearchParams(location.search).get('collection');
  const nav = [
    {
      label: 'All titles',
      icon: LibraryBig,
      to: '/library',
      active: location.pathname === '/library' && !shelf && !collection,
      count: stats.data?.total,
    },
    {
      label: 'In progress',
      icon: BookOpen,
      to: '/library?status=reading',
      active: location.pathname === '/library' && shelf === 'reading',
      count: stats.data?.reading,
    },
    {
      label: 'Up next',
      icon: Bookmark,
      to: '/library?status=planned',
      active: location.pathname === '/library' && shelf === 'planned',
      count: stats.data?.planned,
    },
    {
      label: 'Completed',
      icon: CheckCheck,
      to: '/library?status=completed',
      active: location.pathname === '/library' && shelf === 'completed',
      count: stats.data?.completed,
    },
  ];
  const reading = location.pathname.endsWith('/read');
  const landing = location.pathname === '/';
  const fullscreen = reading || landing;
  return (
    <>
      <a href="#main" className="skip-link">
        Skip to content
      </a>
      {!fullscreen ? (
        <>
          <header className="mobile-header">
            <Link to="/" className="brand">
              <BookOpen size={24} />
              lector
            </Link>
            <button
              className="icon-button"
              aria-label={menu ? 'Close navigation' : 'Open navigation'}
              aria-expanded={menu}
              onClick={() => setMenu(!menu)}
            >
              {menu ? <X /> : <Menu />}
            </button>
          </header>
          <aside className={`sidebar ${menu ? 'is-open' : ''}`}>
            <Link to="/" className="brand">
              <span className="brand-mark">
                <BookOpen size={24} strokeWidth={1.6} />
              </span>
              lector
            </Link>
            <div className="workspace-label">
              <span className="workspace-avatar">M</span>
              <div>
                My workspace<small>Personal library</small>
              </div>
              <span className="local-dot" />
            </div>
            <nav aria-label="Main navigation">
              <p className="nav-label">YOUR LIBRARY</p>
              {nav.map(({ label, icon: Icon, to, active, count }) => (
                <Link
                  onClick={() => setMenu(false)}
                  key={label}
                  to={to}
                  className={`nav-link ${active ? 'active' : ''}`}
                  aria-current={active ? 'page' : undefined}
                >
                  <Icon size={18} strokeWidth={1.7} />
                  <span>{label}</span>
                  {count !== undefined ? <span className="nav-count">{count}</span> : null}
                </Link>
              ))}
              <NavLink onClick={() => setMenu(false)} to="/history" className="nav-link">
                <Clock3 size={18} strokeWidth={1.7} />
                <span>Activity</span>
              </NavLink>
              <p className="nav-label collections-label">
                COLLECTIONS <Folder size={13} />
              </p>
              {stats.data?.collections.map((name) => (
                <Link
                  onClick={() => setMenu(false)}
                  key={name}
                  to={`/library?collection=${encodeURIComponent(name)}`}
                  className={`nav-link collection-link ${collection === name ? 'active' : ''}`}
                >
                  <span className="collection-dot" />
                  {name}
                </Link>
              ))}
              {!stats.data?.collections.length ? (
                <p className="sidebar-hint">
                  Give your titles a collection
                  <br />
                  to create a personal shelf.
                </p>
              ) : null}
              <div className="sidebar-divider" />
              <NavLink onClick={() => setMenu(false)} to="/plugins" className="nav-link">
                <Puzzle size={18} strokeWidth={1.7} />
                <span>Extensions</span>
              </NavLink>
              <NavLink onClick={() => setMenu(false)} to="/settings" className="nav-link">
                <Settings2 size={18} strokeWidth={1.7} />
                <span>Preferences</span>
              </NavLink>
            </nav>
            <div className="sidebar-bottom">
              <div className="local-note">
                <span className="local-dot" />
                <div>
                  Local by nature<small>Your library. Your space.</small>
                </div>
              </div>
              <button className="button sidebar-import" onClick={() => setDialog('import')}>
                <ArrowDownToLine size={16} />
                Import a file
              </button>
            </div>
          </aside>
        </>
      ) : null}
      <main
        id="main"
        className={landing ? 'landing-main' : reading ? 'reader-main' : 'main'}
        tabIndex={-1}
      >
        {!fullscreen ? (
          <div className="topbar">
            <span>
              <LibraryBig size={15} /> Personal space <span className="breadcrumb-slash">/</span>{' '}
              <strong>
                {location.pathname.startsWith('/plugins')
                  ? 'Extensions'
                  : location.pathname === '/history'
                    ? 'Activity'
                    : location.pathname === '/settings'
                      ? 'Preferences'
                      : 'Library'}
              </strong>
            </span>
            <span className="privacy-label">
              <span className="local-dot" />
              Stored locally
            </span>
          </div>
        ) : null}
        <Suspense fallback={<Loading />}>
          <Routes>
            <Route
              path="/"
              element={
                location.search ? (
                  <Navigate to={`/library${location.search}`} replace />
                ) : (
                  <Landing />
                )
              }
            />
            <Route
              path="/library"
              element={
                <Library onAdd={() => setDialog('add')} onImport={() => setDialog('import')} />
              }
            />
            <Route path="/items/:id" element={<Detail />} />
            <Route path="/items/:id/read" element={<Reader />} />
            <Route path="/plugins" element={<Plugins />} />
            <Route path="/history" element={<History />} />
            <Route path="/settings" element={<Settings />} />
            <Route
              path="*"
              element={
                <div className="empty-state">
                  <h1>This page isn’t on the shelf.</h1>
                  <Link to="/library" className="button primary">
                    Back to your library
                  </Link>
                </div>
              }
            />
          </Routes>
        </Suspense>
      </main>
      {dialog === 'add' ? <ItemForm close={() => setDialog(null)} /> : null}
      {dialog === 'import' ? <ImportDialog close={() => setDialog(null)} /> : null}
      {!fullscreen ? (
        <button
          className="mobile-add button primary"
          onClick={() => setDialog('add')}
          aria-label="Add a title"
        >
          <Plus size={22} />
        </button>
      ) : null}
    </>
  );
}
