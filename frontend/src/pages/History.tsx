import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { ArrowUpRight, Clock3 } from 'lucide-react';
import { request } from '../api';
import { statusLabels, type HistoryEvent } from '../types';
import { ErrorMessage, Loading } from '../components/Feedback';
export default function History() {
  const query = useQuery({
    queryKey: ['history'],
    queryFn: () => request<HistoryEvent[]>('/history'),
  });
  return (
    <div className="page">
      <header className="page-heading">
        <div>
          <p className="eyebrow">ONE PAGE AT A TIME</p>
          <h1>
            Your story so far<span className="heading-dot">.</span>
          </h1>
          <p className="page-description">
            The little moments that add up. Your last 100 reading updates.
          </p>
        </div>
      </header>
      <ErrorMessage error={query.error} />
      {query.isPending ? (
        <Loading />
      ) : query.data?.length ? (
        <ol className="activity-list">
          {query.data.map((event) => (
            <li key={event.id}>
              <span className="activity-icon">
                <Clock3 size={18} />
              </span>
              <div>
                <Link to={`/items/${event.item_id}`}>
                  {event.title} <ArrowUpRight size={14} />
                </Link>
                <p>
                  {statusLabels[event.status]} · Progress: {event.current}
                </p>
              </div>
              <time dateTime={event.at}>
                {new Intl.DateTimeFormat(undefined, {
                  dateStyle: 'medium',
                  timeStyle: 'short',
                }).format(new Date(event.at))}
              </time>
            </li>
          ))}
        </ol>
      ) : (
        <div className="empty-state">
          <div className="empty-icon">
            <Clock3 size={30} />
          </div>
          <h2>Every story starts somewhere</h2>
          <p>Update a title’s progress to begin your reading history.</p>
          <Link to="/library" className="button primary">
            Explore your library
          </Link>
        </div>
      )}
    </div>
  );
}
