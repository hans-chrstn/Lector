<script lang="ts">
	import { api } from '$lib/services/api';
	import BaseGridItem from './base/BaseGridItem.svelte';
	import ProgressRing from './ProgressRing.svelte';

	interface Props {
		title: string;
		cover_url: string;
		meta?: string;
		read_chapters?: number;
		total_chapters?: number;
		onclick: () => void;
	}
	let { title, cover_url, meta, read_chapters = 0, total_chapters = 0, onclick }: Props = $props();

	const hasChapters = $derived(total_chapters > 0);
</script>

<BaseGridItem {onclick}>
	{#snippet media()}
		<img
			src={api.getProxyImage(cover_url) || '/placeholder.png'}
			alt={title}
			{title}
			loading="lazy"
			decoding="async"
		/>
	{/snippet}

	{#snippet overlay()}
		{#if hasChapters}
			<div class="progress-ring-box">
				<ProgressRing
					value={read_chapters}
					total={total_chapters}
					size={28}
					stroke={3}
					showText={true}
				/>
			</div>
		{/if}
	{/snippet}

	{#snippet content()}
		<div class="title">{title}</div>
		{#if meta}
			<div class="meta">{meta}</div>
		{/if}
	{/snippet}
</BaseGridItem>

<style>
	.progress-ring-box {
		position: absolute;
		top: 0.5rem;
		right: 0.5rem;
		background: rgba(9, 9, 11, 0.8);
		backdrop-filter: blur(8px);
		padding: 4px;
		border-radius: 50%;
		display: flex;
		z-index: 10;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.title {
		color: var(--text-main);
		font-size: 0.875rem;
		font-weight: 600;
		line-height: 1.25;
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
		margin-bottom: 0.25rem;
	}

	.meta {
		color: var(--text-dim);
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: capitalize;
	}
</style>
