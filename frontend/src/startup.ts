import type { QueryClient } from '@tanstack/react-query';
import { request } from './api';
import type { Page, Stats } from './types';

export const routeModules = {
  Detail: () => import('./pages/Detail'),
  Reader: () => import('./pages/Reader'),
  Plugins: () => import('./pages/Plugins'),
  History: () => import('./pages/History'),
  Settings: () => import('./pages/Settings'),
};

export async function warmUp(client: QueryClient, signal: AbortSignal) {
  const filter = new URLSearchParams({
    q: '',
    kind: '',
    status: '',
    collection: '',
    sort: 'updated',
    offset: '0',
    limit: '24',
  }).toString();
  await Promise.all([
    document.fonts?.ready,
    ...Object.values(routeModules).map((load) => load()),
    client.fetchQuery({
      queryKey: ['stats'],
      queryFn: () => request<Stats>('/stats', { signal }),
      retry: false,
    }),
    client.fetchQuery({
      queryKey: ['items', 'continue'],
      queryFn: () => request<Page>('/items?status=reading&limit=3', { signal }),
      retry: false,
    }),
    client.fetchQuery({
      queryKey: ['items', filter],
      queryFn: () => request<Page>(`/items?${filter}`, { signal }),
      retry: false,
    }),
  ]);
}
