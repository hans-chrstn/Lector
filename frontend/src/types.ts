export type Kind = 'book' | 'manga' | 'anime' | 'video';
export type Status = 'planned' | 'reading' | 'completed' | 'paused';
export interface Item {
  id: string;
  title: string;
  author: string;
  kind: Kind;
  status: Status;
  current: number;
  total: number;
  notes: string;
  collection: string;
  file?: string;
  format?: string;
  created_at: string;
  updated_at: string;
  revision: number;
}
export interface Page {
  items: Item[];
  total: number;
}
export interface Stats {
  total: number;
  reading: number;
  completed: number;
  planned: number;
  collections: string[];
}
export interface HistoryEvent {
  id: number;
  item_id: string;
  title: string;
  current: number;
  status: Status;
  at: string;
}
export interface Plugin {
  id: string;
  name: string;
  description: string;
  version: string;
  hosts: string[] | null;
  ui: boolean;
  actions: { id: string; label: string }[] | null;
  enabled: boolean;
  approved: boolean;
  error?: string;
  generation: string;
}
export interface PluginResult {
  panels: { title: string; text: string; tone: string }[];
  entries: { title: string; author: string; kind: Kind; total: number; text?: string }[];
}
export const kindLabels: Record<Kind, string> = {
  book: 'Books',
  manga: 'Manga',
  anime: 'Anime',
  video: 'Videos',
};
export const statusLabels: Record<Status, string> = {
  planned: 'Up next',
  reading: 'In progress',
  completed: 'Completed',
  paused: 'On hold',
};
export function unit(kind: Kind) {
  return kind === 'anime' || kind === 'video' ? 'episodes' : 'chapters';
}
export function percent(item: Pick<Item, 'current' | 'total'>) {
  return item.total > 0 ? Math.min(100, Math.round((item.current / item.total) * 100)) : 0;
}
