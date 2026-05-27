<script lang="ts">
	import { CheckCircle2, Circle, Loader2 } from 'lucide-svelte';
	import { clsx } from 'clsx';
	import type { Chapter } from '$lib/services/api';

	interface Props {
		chapter: Chapter;
		isCurrent: boolean;
		isMarking: boolean;
		isSelected: boolean;
		onRead: (chapter: Chapter) => void;
		onToggleSelect: (id: number) => void;
	}

	let { chapter, isCurrent, isMarking, isSelected, onRead, onToggleSelect }: Props = $props();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
	class={clsx(
		'chapter-row',
		isCurrent && 'current',
		isSelected && 'selected',
		isMarking && 'marking',
		chapter.is_read && 'read'
	)}
	onclick={() => (isMarking ? onToggleSelect(chapter.id) : onRead(chapter))}
>
	{#if isMarking}
		<div class={clsx('selection-indicator', isSelected && 'active')}>
			{#if isSelected}
				<CheckCircle2 size={18} />
			{:else}
				<Circle size={18} />
			{/if}
		</div>
	{/if}

	<div class="ch-info">
		<span class="ch-num">{chapter.order}</span>
		<span class="ch-title">{chapter.title}</span>
	</div>

	<div class="ch-status">
		{#if chapter.status === 'syncing'}
			<Loader2 size={16} class="spin" />
		{/if}
	</div>
</div>

<style>
	.chapter-row {
		position: relative;
		display: flex;
		align-items: center;
		padding: 0.875rem 1.25rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		cursor: pointer;
		gap: 1rem;
		transition: all 0.2s ease;
	}
	.chapter-row:hover {
		border-color: var(--border-bright);
		background: var(--bg-surface-hover);
		transform: translateX(4px);
	}
	.chapter-row.current {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.05);
	}
	.chapter-row.selected {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}
	.chapter-row.read {
		opacity: 0.5;
	}
	.chapter-row.read .ch-title {
		color: var(--text-dim);
	}
	.selection-indicator {
		color: var(--text-dim);
	}
	.selection-indicator.active {
		color: var(--primary);
	}
	.ch-info {
		flex: 1;
		display: flex;
		gap: 1rem;
		align-items: baseline;
	}
	.ch-num {
		font-size: 0.75rem;
		font-weight: 800;
		opacity: 0.4;
		min-width: 24px;
	}
	.ch-title {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--text-main);
	}
	.ch-status {
		display: flex;
		align-items: center;
	}
	.spin {
		animation: spin 1s linear infinite;
	}
	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
</style>
