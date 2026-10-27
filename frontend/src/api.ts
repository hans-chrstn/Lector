import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Item } from './types';
export async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`/api${path}`, {
    ...init,
    headers: {
      ...(init?.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }),
      ...init?.headers,
    },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.error || `Request failed (${response.status}). Please try again.`);
  }
  return response.status === 204 ? (undefined as T) : response.json();
}
export function useSaveItem() {
  const cache = useQueryClient();
  return useMutation({
    mutationFn: (item: Partial<Item>) =>
      request<Item>(item.id ? `/items/${item.id}` : '/items', {
        method: item.id ? 'PUT' : 'POST',
        body: JSON.stringify(item),
      }),
    onSuccess: () => {
      void cache.invalidateQueries({ queryKey: ['items'] });
      void cache.invalidateQueries({ queryKey: ['item'] });
      void cache.invalidateQueries({ queryKey: ['stats'] });
      void cache.invalidateQueries({ queryKey: ['history'] });
    },
  });
}
