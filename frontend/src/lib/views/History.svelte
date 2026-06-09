<script lang="ts">
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import CheckCircle2 from 'lucide-svelte/icons/check-circle-2';
	import CheckSquare from 'lucide-svelte/icons/check-square';
	import Square from 'lucide-svelte/icons/square';
	import X from 'lucide-svelte/icons/x';
	import Clock from 'lucide-svelte/icons/clock';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import { clsx } from 'clsx';
	import { SvelteSet } from 'svelte/reactivity';
	import type { Document } from '$lib/services/api';

	interface Props {
		history: Document[];
		onOpenDocument: (document: Document) => void;
		onRemove: (id: number) => void;
		onClearAll: () => void;
		onBatchRemove: (ids: number[]) => void;
	}
	let { history, onOpenDocument, onRemove, onClearAll, onBatchRemove }: Props = $props();

	let markMode = $state(false);
	let selectedIds = new SvelteSet<number>();

	function toggleSelect(id: number) {
		if (selectedIds.has(id)) selectedIds.delete(id);
		else selectedIds.add(id);
	}

	function selectAll() {
		if (selectedIds.size === history.length) {
			selectedIds.clear();
		} else {
			for (const doc of history) {
				selectedIds.add(doc.id);
			}
		}
	}

	async function handleBatchDelete() {
		if (selectedIds.size === 0) return;
		await onBatchRemove(Array.from(selectedIds));
		selectedIds.clear();
		markMode = false;
	}
</script>

<BasePage title="History" subtitle="Your recently read titles">
	{#snippet actions()}
		{#if history.length > 0}
			<button
				class={clsx('icon-btn', markMode && 'active')}
				onclick={() => {
					markMode = !markMode;
					if (!markMode) selectedIds.clear();
				}}
				title="Selection Mode"
			>
				<CheckCircle2 size={20} />
			</button>
			<button class="icon-btn danger" onclick={onClearAll} title="Clear All History">
				<Trash2 size={20} />
			</button>
		{/if}
	{/snippet}

	<div class="grid">
		{#each history as document (document.id)}
			<div
				class={clsx(
					'card-wrapper',
					markMode && 'marking',
					selectedIds.has(document.id) && 'selected'
				)}
				role="button"
				tabindex="0"
				onclick={() => {
					if (markMode) toggleSelect(document.id);
					else onOpenDocument(document);
				}}
				onkeydown={(e) => {
					if (e.key === 'Enter' || e.key === ' ') {
						if (markMode) toggleSelect(document.id);
						else onOpenDocument(document);
					}
				}}
			>
				<DocumentGridItem
					title={document.title}
					cover_url={document.cover_url}
					meta={document.source}
					read_chapters={document.read_chapters}
					total_chapters={document.total_chapters}
					onclick={() => {}}
				/>

				{#if markMode}
					<div class="mark-overlay">
						<div class={clsx('mark-indicator', selectedIds.has(document.id) && 'checked')}>
							{#if selectedIds.has(document.id)}
								<CheckCircle2 size={16} />
							{/if}
						</div>
					</div>
				{:else}
					<button
						class="quick-remove"
						onclick={(e) => {
							e.stopPropagation();
							onRemove(document.id);
						}}
						title="Remove from History"
					>
						<X size={16} />
					</button>
				{/if}
			</div>
		{:else}
			<div class="empty-state">
				<div class="empty-icon-box">
					<Clock size={48} class="text-dim" />
				</div>
				<h3>No history yet</h3>
				<p>Your recently read books will appear here for easy access.</p>
			</div>
		{/each}
	</div>

	<div class="history-batch-toolbar" class:visible={markMode && selectedIds.size > 0}>
		<div class="toolbar-content">
			<button class="select-all-btn" onclick={selectAll}>
				{#if selectedIds.size === history.length}
					<CheckSquare size={18} class="text-primary" />
				{:else}
					<Square size={18} />
				{/if}
				<span>{selectedIds.size === history.length ? 'Deselect All' : 'Select All'}</span>
			</button>
			<div class="divider"></div>
			<span class="count-badge">{selectedIds.size} <span class="dim">selected</span></span>
			<div class="spacer"></div>
			<button class="action-btn danger" onclick={handleBatchDelete}>
				<Trash2 size={16} />
				<span>Remove Selected</span>
			</button>
		</div>
	</div>
</BasePage>

<style>
	.icon-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
		width: 40px;
		height: 40px;
		border-radius: 12px;
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
		background: rgba(var(--primary-rgb), 0.1);
	}

	.icon-btn.danger:hover {
		color: #ef4444;
		border-color: #ef444450;
		background: #ef444410;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 2rem;
	}

	.card-wrapper {
		position: relative;
		border-radius: 20px;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.card-wrapper:hover {
		transform: translateY(-4px);
	}

	.mark-overlay {
		position: absolute;
		top: 0.75rem;
		left: 0.75rem;
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
		transition: all 0.2s;
	}

	.mark-indicator.checked {
		background: var(--primary);
		border-color: var(--primary);
	}

	.quick-remove {
		position: absolute;
		top: 0.75rem;
		left: 0.75rem;
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: rgba(9, 9, 11, 0.8);
		backdrop-filter: blur(8px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		opacity: 0;
		transition: all 0.2s;
		z-index: 20;
	}

	.card-wrapper:hover .quick-remove {
		opacity: 1;
	}

	.quick-remove:hover {
		background: #ef4444;
		border-color: #ef4444;
		transform: scale(1.1);
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
		gap: 1.5rem;
	}

	.empty-icon-box {
		width: 80px;
		height: 80px;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.empty-state h3 {
		color: var(--text-main);
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
	}

	.empty-state p {
		max-width: 300px;
		margin: 0;
		font-size: 0.9375rem;
		line-height: 1.6;
	}

	.history-batch-toolbar {
		position: fixed;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%) translateY(100px);
		z-index: 1000;
		background: var(--bg-surface);
		border: 1px solid var(--border-bright);
		padding: 0.5rem;
		border-radius: 100px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		opacity: 0;
		pointer-events: none;
		min-width: 340px;
	}

	.history-batch-toolbar.visible {
		transform: translateX(-50%) translateY(0);
		opacity: 1;
		pointer-events: auto;
	}

	.toolbar-content {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0 0.5rem;
	}

	.select-all-btn {
		background: none;
		border: none;
		color: var(--text-main);
		display: flex;
		align-items: center;
		gap: 0.75rem;
		font-size: 0.8125rem;
		font-weight: 700;
		cursor: pointer;
		padding: 0.5rem 1rem;
		border-radius: 100px;
		transition: all 0.2s;
	}

	.select-all-btn:hover {
		background: var(--bg-surface-hover);
	}

	.divider {
		width: 1px;
		height: 24px;
		background: var(--border-main);
	}

	.count-badge {
		font-size: 0.8125rem;
		font-weight: 700;
		color: var(--text-main);
		white-space: nowrap;
	}

	.count-badge .dim {
		font-weight: 500;
		opacity: 0.5;
	}

	.spacer {
		flex: 1;
	}

	.action-btn {
		background: none;
		border: none;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.8125rem;
		font-weight: 700;
		padding: 0.6rem 1.25rem;
		border-radius: 100px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.action-btn.danger {
		color: #f43f5e;
	}

	.action-btn.danger:hover {
		background: rgba(244, 63, 92, 0.1);
	}

	.text-primary {
		color: var(--primary);
	}
	.text-dim {
		color: var(--text-dim);
	}

	@media (max-width: 768px) {
		.grid {
			grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
			gap: 1rem;
		}
		.history-batch-toolbar {
			bottom: 1rem;
			min-width: calc(100% - 2rem);
		}
	}
</style>
