<script lang="ts">
	import BaseGridItem from './base/BaseGridItem.svelte';
	import ProgressRing from './ProgressRing.svelte';
	import CoverImage from './CoverImage.svelte';

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
		<CoverImage src={cover_url} alt={title} />
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
		<div class="title-meta-stack">
			<div class="title">{title}</div>
			{#if meta}
				<div class="source-badge">{meta}</div>
			{/if}
		</div>
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

	.title-meta-stack {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}

	.title {
		color: var(--text-main);
		font-size: 0.875rem;
		font-weight: 700;
		line-height: 1.25;
		display: -webkit-box;
		line-clamp: 2;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.source-badge {
		display: inline-flex;
		align-items: center;
		padding: 0.15rem 0.5rem;
		background: rgba(var(--primary-rgb), 0.1);
		color: var(--primary);
		border: 1px solid rgba(var(--primary-rgb), 0.2);
		border-radius: 6px;
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		width: fit-content;
	}
</style>
