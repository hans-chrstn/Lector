<script lang="ts">
	import X from 'lucide-svelte/icons/x';
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import Send from 'lucide-svelte/icons/send';

	interface Note {
		id: number;
		content: string;
		created_at: string;
	}

	interface Props {
		notes: Note[];
		newNoteText: string;
		onAddNote: () => void;
		onDeleteNote: (id: number) => void;
		onClose: () => void;
	}

	let { notes, newNoteText = $bindable(), onAddNote, onDeleteNote, onClose }: Props = $props();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="reader-sidebar right-sidebar" onclick={(e) => e.stopPropagation()}>
	<header>
		<span>Notebook</span>
		<button type="button" class="close-sidebar" onclick={onClose}>
			<X size={16} />
		</button>
	</header>
	<div class="notes-container">
		<div class="notes-list">
			{#each notes as n (n.id)}
				<div class="note-item">
					<div class="note-content">{n.content}</div>
					<div class="note-footer">
						<span class="date">{new Date(n.created_at).toLocaleString()}</span>
						<button type="button" onclick={() => onDeleteNote(n.id)}>
							<Trash2 size={12} />
						</button>
					</div>
				</div>
			{:else}
				<div class="empty-list">Your notebook is empty.</div>
			{/each}
		</div>
		<div class="note-input-box">
			<textarea bind:value={newNoteText} placeholder="Write a note..."></textarea>
			<button type="button" class="send-btn" onclick={onAddNote} disabled={!newNoteText.trim()}>
				<Send size={16} />
			</button>
		</div>
	</div>
</div>

<style>
	.reader-sidebar {
		position: fixed;
		top: 0;
		right: 0;
		bottom: 0;
		width: 300px;
		background: var(--surface);
		border-left: 1px solid var(--border-main);
		z-index: 200;
		display: flex;
		flex-direction: column;
		box-shadow: -20px 0 40px rgba(0, 0, 0, 0.2);
		color: var(--text);
	}
	.right-sidebar {
		animation: slideInRight 0.3s ease;
	}
	.reader-sidebar header {
		padding: 1.25rem;
		border-bottom: 1px solid var(--border-main);
		display: flex;
		justify-content: space-between;
		align-items: center;
		color: var(--text-strong);
		font-weight: 700;
		font-size: 0.875rem;
	}
	.notes-container {
		flex: 1;
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
	.notes-list {
		flex: 1;
		overflow-y: auto;
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.note-item {
		background: var(--bg);
		border: 1px solid var(--border-main);
		border-radius: 8px;
		padding: 0.75rem;
	}
	.note-content {
		font-size: 0.875rem;
		color: var(--text-strong);
		line-height: 1.5;
		margin-bottom: 0.5rem;
	}
	.note-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.7rem;
		opacity: 0.5;
	}
	.note-input-box {
		padding: 1rem;
		border-top: 1px solid var(--border-main);
		background: var(--surface);
		display: flex;
		gap: 0.5rem;
	}
	.note-input-box textarea {
		flex: 1;
		background: var(--bg);
		border: 1px solid var(--border-main);
		border-radius: 6px;
		padding: 0.5rem;
		color: var(--text-strong);
		font-size: 0.875rem;
		resize: none;
		height: 60px;
		outline: none;
	}
	.send-btn {
		background: var(--primary);
		color: #fff;
		padding: 0.5rem;
		border-radius: 6px;
		display: flex;
		align-items: center;
		align-self: flex-end;
	}
	.send-btn:disabled {
		opacity: 0.3;
	}
	.empty-list {
		padding: 2rem;
		text-align: center;
		font-size: 0.875rem;
		opacity: 0.5;
	}
	button {
		appearance: none;
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		color: inherit;
		font-family: inherit;
		cursor: pointer;
		outline: none;
	}
	@keyframes slideInRight {
		from {
			transform: translateX(100%);
		}
		to {
			transform: translateX(0);
		}
	}
</style>
