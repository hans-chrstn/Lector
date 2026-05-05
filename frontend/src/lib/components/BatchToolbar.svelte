<script lang="ts">
	import {
		ChevronDown,
		Trash2,
		Archive,
		ArchiveRestore,
		BookOpen,
		BookOpenCheck,
		CheckSquare,
		Square
	} from 'lucide-svelte';

	import { type Group } from '$lib/services/api';

	interface Props {
		selectedCount: number;
		totalCount: number;
		groups: Group[];
		showArchived: boolean;
		onAction: (action: string, param?: any) => void;
		onSelectAll: () => void;
	}
	let { selectedCount, totalCount, groups, showArchived, onAction, onSelectAll }: Props = $props();
</script>

<div class="batch-toolbar" class:visible={selectedCount > 0}>
	<div class="toolbar-content">
		<div class="selection-info">
			<button
				class="select-all-btn"
				onclick={onSelectAll}
				title={selectedCount === totalCount ? 'Deselect All' : 'Select All'}
			>
				{#if selectedCount === totalCount}
					<CheckSquare size={18} class="text-primary" />
				{:else}
					<Square size={18} />
				{/if}
			</button>
			<span class="count-badge">{selectedCount} <span class="dim">selected</span></span>
		</div>

		<div class="divider"></div>

		<div class="actions">
			<div class="group-btn-wrapper">
				<button class="action-btn">
					<span>Move to</span>
					<ChevronDown size={14} />
				</button>
				<div class="dropdown">
					<button onclick={() => onAction('move', 0)}>No Group</button>
					{#each groups as g (g.id)}
						<button onclick={() => onAction('move', g.id)}>{g.name}</button>
					{/each}
				</div>
			</div>

			<div class="group-btn-wrapper">
				<button class="action-btn">
					<span>Export</span>
					<ChevronDown size={14} />
				</button>
				<div class="dropdown">
					<button onclick={() => onAction('export', 'epub')}>EPUB</button>
					<button onclick={() => onAction('export', 'pdf')}>PDF</button>
				</div>
			</div>

			<button class="action-btn" onclick={() => onAction('mark-read')} title="Mark All Read">
				<BookOpenCheck size={16} />
				<span class="label">Read</span>
			</button>

			<button class="action-btn" onclick={() => onAction('mark-unread')} title="Mark All Unread">
				<BookOpen size={16} />
				<span class="label">Unread</span>
			</button>

			<button
				class="action-btn"
				onclick={() => onAction('archive')}
				title={showArchived ? 'Restore from Archive' : 'Archive Selected'}
			>
				{#if showArchived}
					<ArchiveRestore size={16} />
				{:else}
					<Archive size={16} />
				{/if}
				<span class="label">{showArchived ? 'Restore' : 'Archive'}</span>
			</button>

			<button
				class="action-btn danger"
				onclick={() => onAction('delete')}
				title="Delete Permanently"
			>
				<Trash2 size={16} />
				<span class="label">Delete</span>
			</button>
		</div>
	</div>
</div>

<style>
	.batch-toolbar {
		position: fixed;
		bottom: 2rem;
		left: 50%;
		transform: translateX(-50%) translateY(100px);
		z-index: 1000;
		background: var(--bg-surface);
		border: 1px solid var(--border-bright);
		padding: 0.5rem;
		border-radius: 100px;
		box-shadow:
			0 20px 40px rgba(0, 0, 0, 0.4),
			0 0 0 1px rgba(255, 255, 255, 0.05);
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		opacity: 0;
		pointer-events: none;
		max-width: 90vw;
	}

	.batch-toolbar.visible {
		transform: translateX(-50%) translateY(0);
		opacity: 1;
		pointer-events: auto;
	}

	.toolbar-content {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0 0.5rem;
	}

	.selection-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding-left: 0.5rem;
	}

	.select-all-btn {
		background: none;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		display: flex;
		padding: 0.25rem;
		border-radius: 4px;
		transition: all 0.2s;
	}

	.select-all-btn:hover {
		background: rgba(128, 128, 128, 0.1);
		color: var(--text-main);
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

	.divider {
		width: 1px;
		height: 24px;
		background: var(--border-main);
		margin: 0 0.5rem;
	}

	.actions {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.action-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		padding: 0.5rem 0.75rem;
		border-radius: 100px;
		font-size: 0.75rem;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}

	.action-btn:hover {
		background: rgba(128, 128, 128, 0.08);
		color: var(--text-main);
	}
	.action-btn.danger {
		color: #f43f5e;
	}
	.action-btn.danger:hover {
		background: rgba(244, 63, 92, 0.1);
	}

	.group-btn-wrapper {
		position: relative;
	}
	.dropdown {
		position: absolute;
		bottom: 100%;
		left: 50%;
		transform: translateX(-50%) translateY(-10px);
		background: var(--bg-surface);
		border: 1px solid var(--border-bright);
		border-radius: 12px;
		padding: 0.5rem;
		min-width: 140px;
		display: none;
		flex-direction: column;
		gap: 0.25rem;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
		margin-bottom: 0.5rem;
	}

	.group-btn-wrapper:hover .dropdown {
		display: flex;
	}

	.dropdown button {
		background: none;
		border: none;
		color: var(--text-main);
		padding: 0.6rem 0.75rem;
		text-align: left;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
	}

	.dropdown button:hover {
		background: var(--bg-main);
		color: var(--primary);
	}

	@media (max-width: 768px) {
		.label {
			display: none;
		}
		.batch-toolbar {
			bottom: 1rem;
			padding: 0.25rem;
		}
		.action-btn {
			padding: 0.5rem;
		}
	}
</style>
