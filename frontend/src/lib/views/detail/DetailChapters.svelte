<script lang="ts">
	import { CheckCircle2, ListFilter } from 'lucide-svelte';
	import type { Chapter } from '$lib/services/api';
	import { SvelteSet } from 'svelte/reactivity';
	import ChapterListItem from './ChapterListItem.svelte';
	import ChapterBatchToolbar from './ChapterBatchToolbar.svelte';

	interface Props {
		chapters: Chapter[];
		readChapters: number;
		progress: any;
		onReadChapter: (chapter: Chapter) => void;
		onBatchAction: (ids: number[], isRead: boolean) => void;
	}

	let { chapters = [], readChapters, progress, onReadChapter, onBatchAction }: Props = $props();

	let markMode = $state(false);
	let selectedIds = new SvelteSet<number>();

	function toggleSelect(id: number) {
		if (selectedIds.has(id)) selectedIds.delete(id);
		else selectedIds.add(id);
	}

	function selectAll() {
		if (selectedIds.size === chapters.length) selectedIds.clear();
		else chapters.forEach((c) => selectedIds.add(c.id));
	}

	async function runBatch(isRead: boolean) {
		const ids = Array.from(selectedIds);
		await onBatchAction(ids, isRead);
		selectedIds.clear();
		markMode = false;
	}
</script>

<div class="chapters-section">
	<header class="section-header">
		<div class="header-left">
			<h2>Chapters</h2>
			<div class="progress-pill">
				<CheckCircle2 size={14} />
				<span>{readChapters} / {chapters.length}</span>
			</div>
		</div>
		<div class="header-right">
			<button class="icon-btn" class:active={markMode} onclick={() => (markMode = !markMode)}>
				<ListFilter size={18} />
			</button>
		</div>
	</header>

	<div class="chapters-list">
		{#each chapters as ch (ch.id)}
			<ChapterListItem
				chapter={ch}
				isCurrent={progress?.chapter_id === ch.id}
				isMarking={markMode}
				isSelected={selectedIds.has(ch.id)}
				onRead={onReadChapter}
				onToggleSelect={toggleSelect}
			/>
		{/each}
	</div>

	<ChapterBatchToolbar
		selectedCount={selectedIds.size}
		totalCount={chapters.length}
		onAction={runBatch}
		onSelectAll={selectAll}
	/>
</div>

<style>
	.chapters-section {
		margin-top: 4rem;
	}
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}
	.header-left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	h2 {
		font-size: 1.5rem;
		font-weight: 800;
		letter-spacing: -0.02em;
		margin: 0;
	}
	.progress-pill {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
		padding: 0.35rem 0.75rem;
		border-radius: 100px;
		font-size: 0.75rem;
		font-weight: 700;
	}
	.icon-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-dim);
		width: 40px;
		height: 40px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s;
	}
	.icon-btn:hover {
		border-color: var(--border-bright);
		color: var(--text-main);
	}
	.icon-btn.active {
		color: var(--primary);
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}
	.chapters-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
</style>
