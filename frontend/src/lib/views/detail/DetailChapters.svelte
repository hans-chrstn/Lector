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
	let lastSelectedId = $state<number | null>(null);

	function toggleSelect(id: number) {
		if (selectedIds.has(id)) {
			selectedIds.delete(id);
			if (lastSelectedId === id) lastSelectedId = null;
		} else {
			selectedIds.add(id);
			lastSelectedId = id;
		}
	}

	function selectTop() {
		const targetId = lastSelectedId || Array.from(selectedIds)[0];
		if (!targetId) return;
		const idx = chapters.findIndex((c) => c.id === targetId);
		if (idx === -1) return;
		for (let i = 0; i <= idx; i++) {
			selectedIds.add(chapters[i].id);
		}
	}

	function selectBottom() {
		const targetId = lastSelectedId || Array.from(selectedIds)[selectedIds.size - 1];
		if (!targetId) return;
		const idx = chapters.findIndex((c) => c.id === targetId);
		if (idx === -1) return;
		for (let i = idx; i < chapters.length; i++) {
			selectedIds.add(chapters[i].id);
		}
	}

	function selectBetween() {
		const ids = Array.from(selectedIds).sort((a, b) => {
			const idxA = chapters.findIndex((c) => c.id === a);
			const idxB = chapters.findIndex((c) => c.id === b);
			return idxA - idxB;
		});
		if (ids.length < 2) return;
		const startIdx = chapters.findIndex((c) => c.id === ids[0]);
		const endIdx = chapters.findIndex((c) => c.id === ids[ids.length - 1]);
		for (let i = startIdx; i <= endIdx; i++) {
			selectedIds.add(chapters[i].id);
		}
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
		onSelectBetween={selectBetween}
		onSelectTop={selectTop}
		onSelectBottom={selectBottom}
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
		display: grid;
		grid-template-columns: 1fr;
		gap: 0.75rem;
	}

	@media (min-width: 1200px) {
		.chapters-list {
			grid-template-columns: 1fr 1fr;
			gap: 1rem;
		}
	}
</style>
