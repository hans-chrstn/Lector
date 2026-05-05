<script lang="ts">
	import { X, Plus, Trash2 } from 'lucide-svelte';
	import { clsx } from 'clsx';

	interface ChapterMeta {
		id: number;
		title: string;
		order: number;
	}

	interface Bookmark {
		id: number;
		chapter_id: number;
		title: string;
		created_at: string;
	}

	interface Props {
		showSidebar: boolean;
		showBookmarks: boolean;
		chapters: ChapterMeta[];
		bookmarks: Bookmark[];
		currentChapterId: number;
		onReadChapter: (ch: ChapterMeta) => void;
		onAddBookmark: () => void;
		onDeleteBookmark: (id: number) => void;
		onClose: () => void;
	}

	let {
		showSidebar,
		showBookmarks,
		chapters,
		bookmarks,
		currentChapterId,
		onReadChapter,
		onAddBookmark,
		onDeleteBookmark,
		onClose
	}: Props = $props();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="reader-sidebar left-sidebar" onclick={(e) => e.stopPropagation()}>
	<header>
		<span>{showSidebar ? 'Table of Contents' : 'Bookmarks'}</span>
		<div class="sidebar-header-actions">
			{#if showBookmarks}
				<button
					type="button"
					class="add-btn"
					onclick={onAddBookmark}
					title="Add current to bookmarks"
				>
					<Plus size={16} />
				</button>
			{/if}
			<button type="button" class="close-sidebar" onclick={onClose}>
				<X size={16} />
			</button>
		</div>
	</header>
	<div class="chapters-scroll">
		{#if showSidebar}
			{#each chapters as ch (ch.id)}
				<button
					type="button"
					class={clsx('ch-item', ch.id === currentChapterId && 'active')}
					onclick={() => onReadChapter(ch)}
				>
					<span class="num">{ch.order}</span>
					<span class="title">{ch.title}</span>
				</button>
			{/each}
		{:else}
			{#each bookmarks as b (b.id)}
				<div class="bookmark-item">
					<button
						type="button"
						class="b-main"
						onclick={() => {
							const ch = chapters.find((c) => c.id === b.chapter_id);
							if (ch) onReadChapter(ch);
						}}
					>
						<span class="title">{b.title}</span>
						<span class="date">{new Date(b.created_at).toLocaleDateString()}</span>
					</button>
					<button type="button" class="delete-btn" onclick={() => onDeleteBookmark(b.id)}>
						<Trash2 size={14} />
					</button>
				</div>
			{:else}
				<div class="empty-list">No bookmarks yet.</div>
			{/each}
		{/if}
	</div>
</div>

<style>
	.reader-sidebar {
		position: fixed;
		top: 0;
		left: 0;
		bottom: 0;
		width: 300px;
		background: var(--surface);
		border-right: 1px solid var(--border-main);
		z-index: 200;
		display: flex;
		flex-direction: column;
		box-shadow: 20px 0 40px rgba(0, 0, 0, 0.2);
		color: var(--text);
	}
	.left-sidebar {
		animation: slideInLeft 0.3s ease;
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
	.sidebar-header-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}
	.add-btn {
		color: var(--primary);
		padding: 4px;
	}
	.chapters-scroll {
		flex: 1;
		overflow-y: auto;
		padding: 0.5rem;
	}
	.ch-item {
		width: 100%;
		padding: 0.65rem 0.875rem;
		display: flex;
		gap: 0.75rem;
		border-radius: 6px;
		font-size: 0.8125rem;
		color: var(--text);
		text-align: left;
	}
	.ch-item:hover {
		background: rgba(128, 128, 128, 0.08);
		color: var(--text-strong);
	}
	.ch-item.active {
		background: rgba(59, 130, 246, 0.08);
		color: var(--primary);
		font-weight: 600;
	}
	.ch-item .num {
		min-width: 24px;
		opacity: 0.4;
		font-weight: 700;
		font-size: 0.7rem;
	}
	.bookmark-item {
		display: flex;
		align-items: center;
		border-bottom: 1px solid var(--border-main);
	}
	.b-main {
		flex: 1;
		padding: 0.75rem 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		text-align: left;
	}
	.b-main .title {
		font-size: 0.875rem;
		color: var(--text-strong);
	}
	.b-main .date {
		font-size: 0.7rem;
		opacity: 0.5;
	}
	.delete-btn {
		padding: 1rem;
		color: var(--text-dim);
	}
	.delete-btn:hover {
		color: #f43f5e;
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
	@keyframes slideInLeft {
		from {
			transform: translateX(-100%);
		}
		to {
			transform: translateX(0);
		}
	}
</style>
