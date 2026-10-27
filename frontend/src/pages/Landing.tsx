import { ArrowRight, BookOpen, Leaf } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { request } from '../api';
import type { Page } from '../types';
import { Cover } from '../components/Cover';

export default function Landing() {
  const recent = useQuery({
    queryKey: ['items', 'continue'],
    queryFn: () => request<Page>('/items?status=reading&limit=3'),
  });
  const current = recent.data?.items[0];
  return (
    <div className="landing">
      <header className="landing-nav">
        <Link to="/" className="brand">
          <BookOpen size={27} strokeWidth={1.5} />
          lector
        </Link>
        <Link to="/library" className="landing-nav-link">
          Your library <ArrowRight size={16} />
        </Link>
      </header>
      <section className="landing-hero" aria-labelledby="welcome-title">
        <div className="landing-copy">
          <span className="landing-kicker">
            <span className="local-dot" /> A LITTLE ROOM FOR YOURSELF
          </span>
          <h1 id="welcome-title">
            Come in.
            <br />
            Stay for a <em>chapter.</em>
          </h1>
          <p>
            Your stories, a little closer. A quiet place to collect what you love and pick up where
            you left off.
          </p>
          <Link to="/library" className="button primary landing-cta">
            Open your library <ArrowRight size={18} />
          </Link>
          <span className="landing-footnote">No rush. Just one page at a time.</span>
        </div>
        <div className="reading-still-life" aria-hidden="true">
          <div className="reading-halo" />
          <div className="reading-window">
            <span />
            <span />
          </div>
          <div className="reading-table" />
          <div className="resting-book book-bottom">
            <span>THE LITTLE THINGS</span>
          </div>
          <div className="resting-book book-top">
            <span>A place to begin</span>
            <Leaf size={40} strokeWidth={1} />
            <small>THE PERSONAL COLLECTION</small>
          </div>
          <div className="reading-cup">
            <span />
          </div>
          <div className="reading-bookmark" />
          <span className="still-life-caption">a moment, all yours.</span>
        </div>
      </section>
      <section className="landing-return" aria-label="Your next chapter">
        <div className="landing-return-intro">
          <Leaf size={22} strokeWidth={1.4} />
          <div>
            <h2>{current ? 'Right where you left off.' : 'Make yourself at home.'}</h2>
            <p>
              {current
                ? 'A familiar story is waiting for you.'
                : 'Your next favorite belongs here.'}
            </p>
          </div>
        </div>
        {current ? (
          <Link className="landing-current" to={`/items/${current.id}`}>
            <Cover item={current} compact />
            <span>
              <strong>{current.title}</strong>
              <small>{current.author || 'Continue your story'}</small>
            </span>
            <ArrowRight size={20} />
          </Link>
        ) : (
          <Link to="/library" className="landing-empty-link">
            Find a place for your first story <ArrowRight size={18} />
          </Link>
        )}
      </section>
      <footer className="landing-footer">
        <span>Books, manga, and everything in between.</span>
        <span>
          <span className="local-dot" /> Yours. Stored locally.
        </span>
      </footer>
    </div>
  );
}
