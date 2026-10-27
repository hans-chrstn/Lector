import { useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, ArrowRight, Check, Minus, Plus } from 'lucide-react';
import { request, useSaveItem } from '../api';
import type { Item } from '../types';
import { ErrorMessage, Loading } from '../components/Feedback';
interface Content {
  chapters: { title: string; text?: string; image?: number }[];
}
function TextReader({ item }: { item: Item }) {
  const [index, setIndex] = useState(
    Math.max(0, Math.min(item.current, Math.max(0, item.total - 1))),
  );
  const [size, setSize] = useState(19);
  const [theme, setTheme] = useState('paper');
  const save = useSaveItem();
  const query = useQuery({
    queryKey: ['content', item.id],
    queryFn: () => request<Content>(`/items/${item.id}/content`),
    staleTime: Infinity,
  });
  if (query.isPending) return <Loading />;
  if (query.error) return <ErrorMessage error={query.error} />;
  const chapters = query.data!.chapters;
  const chapter = chapters[index];
  async function advance() {
    try {
      await save.mutateAsync({
        ...item,
        current: index + 1,
        status: index + 1 >= chapters.length ? 'completed' : 'reading',
      });
      if (index + 1 < chapters.length) {
        setIndex(index + 1);
        window.scrollTo({ top: 0 });
      }
    } catch {
      return;
    }
  }
  return (
    <div className={`reading-surface theme-${theme}`}>
      <div className="reader-toolbar">
        <label>
          Chapter
          <select
            aria-label="Choose chapter"
            value={index}
            onChange={(e) => setIndex(Number(e.target.value))}
          >
            {chapters.map((c, i) => (
              <option key={i} value={i}>
                {c.title}
              </option>
            ))}
          </select>
        </label>
        <div className="reader-options">
          <button
            className="icon-button"
            aria-label="Decrease text size"
            disabled={size <= 14}
            onClick={() => setSize(size - 1)}
          >
            <Minus size={16} />
          </button>
          <span>{size}</span>
          <button
            className="icon-button"
            aria-label="Increase text size"
            disabled={size >= 30}
            onClick={() => setSize(size + 1)}
          >
            <Plus size={16} />
          </button>
          <select
            aria-label="Reader theme"
            value={theme}
            onChange={(e) => setTheme(e.target.value)}
          >
            <option value="paper">Paper</option>
            <option value="sepia">Sepia</option>
            <option value="night">Night</option>
          </select>
        </div>
      </div>
      <article className="reader-content" style={{ fontSize: size }}>
        <h1>{chapter?.title}</h1>
        {chapter?.image ? (
          <img
            src={`/api/items/${item.id}/pages/${chapter.image}`}
            alt={`Page ${chapter.image} of ${item.title}`}
            className="comic-page"
            width={900}
            height={1300}
          />
        ) : (
          <div className="reading-text">
            {chapter?.text?.split(/\n\s*\n/).map((paragraph, i) => (
              <p key={i}>{paragraph}</p>
            ))}
          </div>
        )}
      </article>
      <footer className="reader-footer">
        <button
          className="button secondary"
          disabled={index === 0}
          onClick={() => {
            setIndex(index - 1);
            window.scrollTo({ top: 0 });
          }}
        >
          <ArrowLeft size={16} />
          Previous
        </button>
        <span>
          {index + 1} / {chapters.length}
        </span>
        <button className="button primary" disabled={save.isPending} onClick={() => void advance()}>
          {save.isPending
            ? 'Saving…'
            : index + 1 === chapters.length
              ? 'Finish reading'
              : 'Read & continue'}
          {index + 1 === chapters.length ? <Check size={16} /> : <ArrowRight size={16} />}
        </button>
      </footer>
      <ErrorMessage error={save.error} />
    </div>
  );
}
export default function Reader() {
  const { id } = useParams();
  const query = useQuery({ queryKey: ['item', id], queryFn: () => request<Item>(`/items/${id}`) });
  const save = useSaveItem();
  if (query.isPending) return <Loading />;
  if (query.error) return <ErrorMessage error={query.error} />;
  const item = query.data!;
  return (
    <div className="reader-shell">
      <header className="reader-header">
        <Link to={`/items/${id}`} className="back-link">
          <ArrowLeft size={17} />
          Library
        </Link>
        <span>{item.title}</span>
        <span className="reader-wordmark">lector</span>
      </header>
      {!item.file ? (
        <div className="empty-state">
          <h1>No file attached</h1>
          <p>This title is a tracking entry. Import a local file to read it here.</p>
        </div>
      ) : ['pdf', 'mp4', 'webm'].includes(item.format || '') ? (
        <div className="media-reader">
          {item.format === 'pdf' ? (
            <iframe src={`/api/items/${id}/file`} title={item.title} className="pdf-viewer" />
          ) : (
            <video
              controls
              preload="metadata"
              src={`/api/items/${id}/file`}
              aria-label={item.title}
            >
              <track kind="captions" />
              <p>Your browser does not support this video format.</p>
            </video>
          )}
          <div className="media-actions">
            <a
              href={`/api/items/${id}/file`}
              target="_blank"
              rel="noreferrer"
              className="button secondary"
            >
              Open original file
            </a>
            <button
              className="button primary"
              disabled={save.isPending || item.status === 'completed'}
              onClick={() =>
                save.mutate({ ...item, current: item.total || item.current, status: 'completed' })
              }
            >
              <Check size={16} />
              Mark complete
            </button>
          </div>
          <ErrorMessage error={save.error} />
        </div>
      ) : (
        <TextReader item={item} />
      )}
    </div>
  );
}
