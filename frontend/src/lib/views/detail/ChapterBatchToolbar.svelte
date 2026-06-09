<script lang="ts">
	import CheckSquare from 'lucide-svelte/icons/check-square';
	import Square from 'lucide-svelte/icons/square';
	import BookOpen from 'lucide-svelte/icons/book-open';
	import BookOpenCheck from 'lucide-svelte/icons/book-open-check';

	interface Props {
		selectedCount: number;
		totalCount: number;
		onAction: (isRead: boolean) => void;
		onSelectAll: () => void;
		onSelectBetween: () => void;
		onSelectTop: () => void;
		onSelectBottom: () => void;
	}

	let {
		selectedCount,
		totalCount,
		onAction,
		onSelectAll,
		onSelectBetween,
		onSelectTop,
		onSelectBottom
	}: Props = $props();
</script>

<div class="batch-toolbar" class:visible={selectedCount > 0}>
	<div class="content">
		<button class="select-all" onclick={onSelectAll}>
			{#if selectedCount === totalCount}
				<CheckSquare size={18} />
			{:else}
				<Square size={18} />
			{/if}
			<span>{selectedCount} Selected</span>
		</button>

		<div class="divider"></div>

		<div class="select-options">
			<button class="action-btn" onclick={onSelectTop}>
				<span>Top</span>
			</button>
			<button class="action-btn" onclick={onSelectBetween} disabled={selectedCount < 2}>
				<span>Between</span>
			</button>
			<button class="action-btn" onclick={onSelectBottom}>
				<span>Bottom</span>
			</button>
		</div>

		<div class="divider"></div>

		<div class="actions">
			<button class="action-btn" onclick={() => onAction(true)}>
				<BookOpenCheck size={18} />
				<span>Mark Read</span>
			</button>
			<button class="action-btn" onclick={() => onAction(false)}>
				<BookOpen size={18} />
				<span>Mark Unread</span>
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
		padding: 0.5rem 1rem;
		border-radius: 100px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		opacity: 0;
		pointer-events: none;
	}
	.batch-toolbar.visible {
		transform: translateX(-50%) translateY(0);
		opacity: 1;
		pointer-events: auto;
	}
	.content {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.select-all {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background: none;
		border: none;
		color: var(--text-main);
		font-weight: 700;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
	}
	.divider {
		width: 1px;
		height: 20px;
		background: var(--border-main);
	}
	.select-options {
		display: flex;
		gap: 0.25rem;
	}
	.actions {
		display: flex;
		gap: 0.5rem;
	}
	.action-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-dim);
		font-size: 0.8125rem;
		font-weight: 600;
		padding: 0.5rem 0.75rem;
		border-radius: 100px;
		cursor: pointer;
		transition: all 0.2s;
	}
	.action-btn:hover:not(:disabled) {
		background: rgba(128, 128, 128, 0.1);
		color: var(--text-main);
	}
	.action-btn:disabled {
		opacity: 0.3;
		cursor: not-allowed;
	}
</style>
