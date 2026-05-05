<script lang="ts">
	import { api } from '$lib/services/api';
	import BaseListItem from './base/BaseListItem.svelte';
	import ProgressRing from './ProgressRing.svelte';
	import { FileText, Globe } from 'lucide-svelte';

	interface Props {
		title: string;
		cover_url: string;
		meta?: string;
		is_local?: boolean;
		read_chapters?: number;
		total_chapters?: number;
		onclick: () => void;
	}
	let {
		title,
		cover_url,
		meta,
		is_local = false,
		read_chapters = 0,
		total_chapters = 0,
		onclick
	}: Props = $props();

	const hasChapters = $derived(total_chapters > 0);
</script>

<BaseListItem {onclick}>
	{#snippet media()}
		<img src={api.getProxyImage(cover_url) || '/placeholder.png'} alt={title} loading="lazy" />
	{/snippet}

	{#snippet content()}
		<div class="info-box">
			<div class="title">{title}</div>
			<div class="meta-row">
				<div class="status-badge" class:local={is_local}>
					{#if is_local}
						<FileText size={12} />
						<span>Local</span>
					{:else if meta}
						<Globe size={12} />
						<span>{meta}</span>
					{/if}
				</div>
				{#if hasChapters}
					<span class="progress-text">
						{read_chapters} / {total_chapters} chapters read
					</span>
				{/if}
			</div>
		</div>
	{/snippet}

	{#snippet actions()}
		{#if hasChapters}
			<div class="progress-box">
				<ProgressRing
					value={read_chapters}
					total={total_chapters}
					size={36}
					stroke={4}
					showText={true}
				/>
			</div>
		{/if}
	{/snippet}
</BaseListItem>

<style>
	.info-box {
		flex: 1;
		min-width: 0;
	}

	.title {
		color: var(--text-main);
		font-size: 1rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.meta-row {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.status-badge {
		display: flex;
		align-items: center;
		gap: 0.35rem;
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.2rem 0.5rem;
		border-radius: 4px;
		background: rgba(var(--primary-rgb), 0.1);
		color: var(--primary);
	}

	.status-badge.local {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}

	.progress-text {
		color: var(--text-dim);
		font-size: 0.75rem;
		font-weight: 500;
	}

	.progress-box {
		flex-shrink: 0;
	}
</style>
