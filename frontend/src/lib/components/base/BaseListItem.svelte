<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronRight } from 'lucide-svelte';

	interface Props {
		onclick: () => void;
		media?: Snippet;
		content: Snippet;
		actions?: Snippet;
		showChevron?: boolean;
	}
	let { onclick, media, content, actions, showChevron = true }: Props = $props();
</script>

<button class="base-list-item" {onclick}>
	<div class="item-inner">
		{#if media}
			<div class="media-slot">
				{@render media()}
			</div>
		{/if}

		<div class="content-slot">
			{@render content()}
		</div>

		{#if actions}
			<div class="actions-slot">
				{@render actions()}
			</div>
		{/if}

		{#if showChevron}
			<div class="chevron-slot">
				<ChevronRight size={20} />
			</div>
		{/if}
	</div>
</button>

<style>
	.base-list-item {
		width: 100%;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 0.75rem 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		text-align: left;
		margin-bottom: 0.75rem;
		display: block;
		outline: none;
	}

	.base-list-item:hover {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
		transform: translateX(4px);
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
	}

	.item-inner {
		display: flex;
		align-items: center;
		gap: 1.25rem;
	}

	.media-slot {
		width: 48px;
		height: 72px;
		border-radius: 6px;
		overflow: hidden;
		flex-shrink: 0;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
	}

	:global(.media-slot img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.content-slot {
		flex: 1;
		min-width: 0;
	}

	.actions-slot {
		flex-shrink: 0;
	}

	.chevron-slot {
		color: var(--text-dim);
		opacity: 0.5;
		margin-left: 0.5rem;
	}

	.base-list-item:hover .chevron-slot {
		opacity: 1;
		color: var(--primary);
	}
</style>
