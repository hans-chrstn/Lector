import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ItemForm } from './ItemForm';
function setup() {
  const close = vi.fn();
  render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { mutations: { retry: false } } })}
    >
      <ItemForm close={close} />
    </QueryClientProvider>,
  );
  return { close, user: userEvent.setup() };
}
describe('adding a title', () => {
  it('keeps the dialog open when discarding an unsaved edit is declined', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(false);
    const { close, user } = setup();
    await user.type(screen.getByLabelText('Title'), 'Still writing');
    const cancel = new Event('cancel', { cancelable: true });
    fireEvent(screen.getByRole('dialog'), cancel);
    expect(cancel.defaultPrevented).toBe(true);
    expect(close).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog')).toHaveAttribute('open');
    expect(screen.getByLabelText('Title')).toHaveValue('Still writing');
  });
  it('submits the selected format and shelf, then closes after persistence', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(new Response(JSON.stringify({ id: 'new' }), { status: 201 }));
    vi.stubGlobal('fetch', fetch);
    const { close, user } = setup();
    await user.type(screen.getByLabelText('Title'), 'A favorite story');
    await user.type(screen.getByLabelText('Author or creator'), 'Ada');
    await user.selectOptions(screen.getByLabelText('Format'), 'manga');
    await user.selectOptions(screen.getByLabelText('Shelf'), 'reading');
    await user.click(screen.getByRole('button', { name: 'Add to library' }));
    await waitFor(() => expect(close).toHaveBeenCalledOnce());
    const body = JSON.parse(fetch.mock.calls[0][1].body);
    expect(body).toMatchObject({
      title: 'A favorite story',
      author: 'Ada',
      kind: 'manga',
      status: 'reading',
      current: 0,
      total: 0,
    });
  });
  it('keeps entered data visible when saving fails', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          new Response(JSON.stringify({ error: 'Storage unavailable' }), { status: 500 }),
        ),
    );
    const { close, user } = setup();
    await user.type(screen.getByLabelText('Title'), 'Keep this title');
    await user.click(screen.getByRole('button', { name: 'Add to library' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('Storage unavailable');
    expect(screen.getByLabelText('Title')).toHaveValue('Keep this title');
    expect(close).not.toHaveBeenCalled();
  });
});
