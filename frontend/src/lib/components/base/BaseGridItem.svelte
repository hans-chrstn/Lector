<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		onclick: () => void;
		media: Snippet;
		overlay?: Snippet;
		content?: Snippet;
		aspectRatio?: string;
	}
	let { onclick, media, overlay, content, aspectRatio = '2/3' }: Props = $props();
</script>

<button class="base-grid-item" {onclick}>
	<div class="card-inner" style="aspect-ratio: {aspectRatio}">
		<div class="media-layer">
			{@render media()}
		</div>

		{#if overlay}
			<div class="overlay-layer">
				{@render overlay()}
			</div>
		{/if}

		{#if content}
			<div class="content-layer">
				{@render content()}
			</div>
		{/if}
	</div>
</button>

<style>
	.base-grid-item {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		width: 100%;
		text-align: left;
		outline: none;
		display: block;
	}

	.card-inner {
		position: relative;
		width: 100%;
		background: var(--bg-surface);
		overflow: hidden;
		border-radius: 12px;
		border: 1px solid var(--border-main);
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.media-layer {
		width: 100%;
		height: 100%;
		z-index: 1;
	}

	:global(.media-layer img) {
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center top;
		transition: transform 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		image-rendering: -webkit-optimize-contrast;
	}

	.overlay-layer {
		position: absolute;
		inset: 0;
		z-index: 10;
		pointer-events: none;
	}

	.content-layer {
		position: absolute;
		inset: 0;
		background: linear-gradient(
			to top,
			rgba(9, 9, 11, 0.95) 0%,
			rgba(9, 9, 11, 0.4) 40%,
			transparent 100%
		);
		display: flex;
		flex-direction: column;
		justify-content: flex-end;
		padding: 1rem;
		z-index: 5;
	}

	@media (hover: hover) {
		.base-grid-item:hover .card-inner {
			border-color: var(--border-bright);
			transform: translateY(-4px);
			box-shadow: 0 12px 24px -8px rgba(0, 0, 0, 0.5);
		}

		.base-grid-item:hover :global(.media-layer img) {
			transform: scale(1.05);
		}
	}
</style>
