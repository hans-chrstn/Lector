<script lang="ts">
	import { X } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import { fade, scale } from 'svelte/transition';

	interface Props {
		show: boolean;
		title?: string;
		onClose: () => void;
		children: Snippet;
		width?: string;
	}

	let { show, title, onClose, children, width = '500px' }: Props = $props();

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape' && show) onClose();
	}
</script>

<svelte:window onkeydown={handleKeydown} />

{#if show}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="modal-backdrop" onclick={onClose} transition:fade={{ duration: 200 }}>
		<div
			class="modal-wrapper"
			onclick={(e) => e.stopPropagation()}
			style="--modal-width: {width}"
			transition:scale={{ duration: 250, start: 0.95, opacity: 0 }}
		>
			<header class="modal-header">
				{#if title}
					<h3>{title}</h3>
				{/if}
				<button class="close-btn" onclick={onClose} aria-label="Close modal">
					<X size={20} />
				</button>
			</header>

			<div class="modal-content">
				{@render children()}
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.7);
		backdrop-filter: blur(8px);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 3000;
		padding: 2rem;
	}

	.modal-wrapper {
		background: var(--bg-surface);
		border: 1px solid var(--border-bright);
		border-radius: 20px;
		width: 100%;
		max-width: var(--modal-width);
		box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		padding: 1.5rem 1.5rem 1rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.modal-header h3 {
		margin: 0;
		font-size: 1.25rem;
		font-weight: 700;
		letter-spacing: -0.02em;
	}

	.close-btn {
		background: none;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 50%;
		display: flex;
		transition: all 0.2s;
	}

	.close-btn:hover {
		color: var(--text-main);
		background: var(--bg-surface-hover);
	}

	.modal-content {
		padding: 0 1.5rem 1.5rem;
	}
</style>
