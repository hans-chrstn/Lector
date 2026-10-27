import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import Plugins from './Plugins';
import type { Plugin } from '../types';
const plugin: Plugin = {
  id: 'notes',
  name: 'Notes',
  description: 'Personal panels',
  version: '1',
  hosts: ['example.org'],
  ui: true,
  actions: [{ id: 'run', label: 'Show notes' }],
  enabled: false,
  approved: false,
  generation: 'v1',
};
function mount() {
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <Plugins />
    </QueryClientProvider>,
  );
  return userEvent.setup();
}
describe('extension permission boundary', () => {
  it('requires an explicit permission review before enabling', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify([plugin]), { status: 200 }));
    fetch.mockImplementation(async () => new Response(JSON.stringify([plugin]), { status: 200 }));
    vi.stubGlobal('fetch', fetch);
    const user = mount();
    await user.click(await screen.findByRole('button', { name: 'Review & enable' }));
    expect(screen.getByRole('dialog')).toHaveTextContent('example.org');
    expect(fetch.mock.calls.every((call) => !call[1]?.method)).toBe(true);
    await user.click(screen.getByRole('button', { name: 'Approve & enable' }));
    await waitFor(() =>
      expect(fetch.mock.calls.some((call) => call[1]?.method === 'PUT')).toBe(true),
    );
    const write = fetch.mock.calls.find((call) => call[1]?.method === 'PUT');
    expect(JSON.parse(write![1].body)).toEqual({ enabled: true, generation: 'v1' });
  });
  it('renders plugin markup as inert text', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation(
        async (path: string) =>
          new Response(
            JSON.stringify(
              path.includes('/actions/')
                ? {
                    panels: [
                      { title: 'A panel', text: '<img src=x onerror=alert(1)>', tone: 'accent' },
                    ],
                    entries: [],
                  }
                : [{ ...plugin, enabled: true }],
            ),
            { status: 200 },
          ),
      ),
    );
    const user = mount();
    await user.click(await screen.findByRole('button', { name: 'Show notes' }));
    expect(await screen.findByText('<img src=x onerror=alert(1)>')).toBeVisible();
    expect(document.querySelector('.extension-panel img')).toBeNull();
  });
});
