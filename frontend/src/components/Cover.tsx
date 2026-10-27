import { BookOpen, Clapperboard, Film, Layers } from 'lucide-react';
import type { Item } from '../types';
const icons = { book: BookOpen, manga: Layers, anime: Clapperboard, video: Film };
export function Cover({ item, compact = false }: { item: Item; compact?: boolean }) {
  const Icon = icons[item.kind];
  const palette = [...item.title].reduce((sum, c) => sum + c.charCodeAt(0), 0) % 6;
  return (
    <div className={`cover palette-${palette} ${compact ? 'compact' : ''}`} aria-hidden="true">
      <div className="cover-line" />
      <span className="cover-kind">
        {item.kind === 'book' ? 'THE READING COLLECTION' : item.kind.toUpperCase()}
      </span>
      <div className="cover-orbit">
        <Icon strokeWidth={1} />
      </div>
      <span className="cover-title">{item.title}</span>
      <span className="cover-author">{item.author || 'YOUR PERSONAL EDITION'}</span>
    </div>
  );
}
