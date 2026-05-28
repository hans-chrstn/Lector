<script lang="ts">
	import { FolderPlus, CheckCircle2, Filter, Grid3X3, Upload, Ghost } from 'lucide-svelte';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';
	import BatchToolbar from '../components/BatchToolbar.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import { clsx } from 'clsx';
	import { SvelteSet } from 'svelte/reactivity';
	import type { Document, Group } from '$lib/services/api';

	interface Props {
		documents: Document[];
		groups: Group[];
		onOpenDocument: (document: Document) => void;
		onCreateGroup: () => void;
		onUpload: (file: File) => void;
		onRefresh: (archived?: boolean) => void;
		onBatchDelete: (ids: number[]) => void;
		onBatchMove: (ids: number[], groupId: number) => void;
		onBatchArchive: (ids: number[], archive: boolean) => void;
		onBatchMarkRead: (ids: number[], isRead: boolean) => void;
	}
	let {
		documents = [],
		groups,
		onOpenDocument,
		onCreateGroup,
		onUpload,
		onRefresh,
		onBatchDelete,
		onBatchMove,
		onBatchArchive,
		onBatchMarkRead
	}: Props = $props();

	let selectedGroupId = $state(0);
	let markMode = $state(false);
	let showArchived = $state(false);
	let selectedIds = new SvelteSet<number>();
	let fileInput: HTMLInputElement;

	let filteredDocuments = $derived(
		(documents || []).filter((n) => selectedGroupId === 0 || n.group_id === selectedGroupId)
	);

	function toggleSelect(id: number) {
		if (selectedIds.has(id)) selectedIds.delete(id);
		else selectedIds.add(id);
	}

	function selectAll() {
		if (selectedIds.size === documents.length) {
			selectedIds.clear();
		} else {
			for (const doc of documents) {
				selectedIds.add(doc.id);
			}
		}
	}

	async function runBatch(action: string, param?: any) {
		const ids = Array.from(selectedIds);
		if (action === 'delete') {
			await onBatchDelete(ids);
		} else if (action === 'move') {
			await onBatchMove(ids, param);
		} else if (action === 'archive') {
			await onBatchArchive(ids, !showArchived);
		} else if (action === 'mark-read') {
			await onBatchMarkRead(ids, true);
		} else if (action === 'mark-unread') {
			await onBatchMarkRead(ids, false);
		}

		selectedIds.clear();
		markMode = false;
		onRefresh(showArchived);
	}

	function handleToggleArchived() {
		showArchived = !showArchived;
		selectedGroupId = 0;
		onRefresh(showArchived);
	}
</script>

<BasePage title="Library" subtitle="Manage and organize your collection">
	{#snippet actions()}
		<input
			type="file"
			accept=".epub"
			style="display: none"
			bind:this={fileInput}
			onchange={(e) => e.currentTarget.files && onUpload(e.currentTarget.files[0])}
		/>
		<button class="icon-btn" onclick={() => fileInput.click()} title="Upload Local Book">
			<Upload size={20} />
		</button>
		<button
			class={clsx('icon-btn', markMode && 'active')}
			onclick={() => (markMode = !markMode)}
			title="Selection Mode"
		>
			<CheckCircle2 size={20} />
		</button>
		<button class="icon-btn" onclick={onCreateGroup} title="New Group">
			<FolderPlus size={20} />
		</button>
	{/snippet}

	{#snippet extraHeader()}
		<div class="filter-bar">
			<div class="tabs">
				<button
					class={clsx('tab', selectedGroupId === 0 && !showArchived && 'active')}
					onclick={() => {
						selectedGroupId = 0;
						showArchived = false;
						onRefresh(false);
					}}
				>
					All Titles
				</button>
				<button class={clsx('tab', showArchived && 'active')} onclick={handleToggleArchived}>
					Archived
				</button>
				{#each groups as g (g.id)}
					<button
						class={clsx('tab', selectedGroupId === g.id && !showArchived && 'active')}
						onclick={() => {
							showArchived = false;
							selectedGroupId = g.id;
							onRefresh(false);
						}}
					>
						{g.name}
					</button>
				{/each}
			</div>

			<div class="display-opts">
				<button class="icon-btn-small"><Grid3X3 size={16} /></button>
				<button class="icon-btn-small"><Filter size={16} /></button>
			</div>
		</div>

		<BatchToolbar
			selectedCount={selectedIds.size}
			totalCount={documents.length}
			{groups}
			{showArchived}
			onAction={runBatch}
			onSelectAll={selectAll}
		/>
	{/snippet}

	<div class="grid">
		{#each filteredDocuments as document (document.id)}
			<div
				class={clsx(
					'card-wrapper',
					markMode && 'marking',
					selectedIds.has(document.id) && 'selected'
				)}
				role="button"
				tabindex="0"
				onclick={(e) => {
					if (markMode) {
						e.preventDefault();
						e.stopPropagation();
						toggleSelect(document.id);
					}
				}}
				onkeydown={(e) => {
					if (markMode && (e.key === 'Enter' || e.key === ' ')) {
						e.preventDefault();
						e.stopPropagation();
						toggleSelect(document.id);
					}
				}}
			>
				<DocumentGridItem
					title={document.title}
					cover_url={document.cover_url}
					meta={document.source}
					read_chapters={document.read_chapters}
					total_chapters={document.total_chapters}
					onclick={() => !markMode && onOpenDocument(document)}
				/>
				{#if markMode}
					<div class="mark-overlay">
						<div class={clsx('mark-indicator', selectedIds.has(document.id) && 'checked')}>
							{#if selectedIds.has(document.id)}
								<CheckCircle2 size={16} />
							{/if}
						</div>
					</div>
				{/if}
			</div>
		{:else}
			<div class="empty-state">
				<Ghost size={40} class="empty-icon" />
				<h3>Your library is empty</h3>
				<p>Upload a file or start searching to build your collection.</p>
			</div>
		{/each}
	</div>
</BasePage>

<style>
	.icon-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
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
		background: var(--bg-surface-hover);
	}

	.icon-btn.active {
		color: var(--primary);
		border-color: var(--primary);
		background: rgba(59, 130, 246, 0.1);
	}

	.filter-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		border-bottom: 1px solid var(--border-main);
		padding-bottom: 0.5rem;
	}

	.tabs {
		display: flex;
		gap: 1.5rem;
	}

	.tab {
		background: none;
		border: none;
		color: var(--text-dim);
		font-size: 0.875rem;
		font-weight: 600;
		padding: 0.5rem 0;
		cursor: pointer;
		position: relative;
		transition: color 0.2s;
	}

	.tab:hover {
		color: var(--text-main);
	}

	.tab.active {
		color: var(--primary);
	}

	.tab.active::after {
		content: '';
		position: absolute;
		bottom: -0.5rem;
		left: 0;
		right: 0;
		height: 2px;
		background: var(--primary);
		border-radius: 2px;
	}

	.display-opts {
		display: flex;
		gap: 0.5rem;
	}

	.icon-btn-small {
		background: none;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		padding: 0.25rem;
	}

	.icon-btn-small:hover {
		color: var(--text-main);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 2rem;
	}

	.card-wrapper {
		position: relative;
	}

	.mark-overlay {
		position: absolute;
		top: 0.5rem;
		left: 0.5rem;
		z-index: 20;
	}

	.mark-indicator {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(4px);
		border: 2px solid white;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
	}

	.mark-indicator.checked {
		background: var(--primary);
		border-color: var(--primary);
	}

	.empty-state {
		grid-column: 1 / -1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8rem 0;
		text-align: center;
		color: var(--text-dim);
		gap: 1rem;
	}

	.empty-state h3 {
		color: var(--text-main);
		margin: 0;
	}

	@media (max-width: 768px) {
		.grid {
			grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
			gap: 1rem;
		}
	}
</style>
