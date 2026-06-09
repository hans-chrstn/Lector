<script lang="ts">
	import { onMount } from 'svelte';
	import { api, type SearchItem } from '$lib/services/api';
	import { clsx } from 'clsx';
	import Compass from 'lucide-svelte/icons/compass';
	import LayoutGrid from 'lucide-svelte/icons/layout-grid';
	import List from 'lucide-svelte/icons/list';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import BasePage from '../components/base/BasePage.svelte';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';

	interface Props {
		pluginName: string;
		tabId: string;
		onSelectDocument: (url: string, source: string) => void;
	}
	let { pluginName, tabId, onSelectDocument }: Props = $props();

	let loading = $state(false);
	let content = $state<SearchItem[]>([]);
	let layout = $state<'grid' | 'list'>('grid');

	async function refresh() {
		loading = true;
		try {
			if (tabId === 'discovery' || tabId === 'system:discovery') {
				const results = await api.getDocumentPopular(pluginName);
				content = results || [];
			} else {
				content = [];
			}
		} finally {
			loading = false;
		}
	}

	onMount(refresh);

	$effect(() => {
		if (pluginName || tabId) {
			refresh();
		}
	});
</script>

<BasePage
	title={pluginName}
	titleBadge={tabId.replace('system:', '')}
	subtitle="Explore content from this source"
	containerClass="capitalize"
>
	{#snippet actions()}
		<div class="actions-group">
			<button class="icon-btn" onclick={refresh} disabled={loading} title="Refresh">
				<RefreshCw size={18} class={clsx(loading && 'animate-spin')} />
			</button>
			<div class="divider"></div>
			<button
				class={clsx('icon-btn', layout === 'grid' && 'active')}
				onclick={() => (layout = 'grid')}
				title="Grid View"
			>
				<LayoutGrid size={18} />
			</button>
			<button
				class={clsx('icon-btn', layout === 'list' && 'active')}
				onclick={() => (layout = 'list')}
				title="List View"
			>
				<List size={18} />
			</button>
		</div>
	{/snippet}

	<div class="explorer-content">
		{#if loading && content.length === 0}
			<div class="loading-state">
				<div class="spinner"></div>
				<span>Loading content...</span>
			</div>
		{:else if content.length === 0}
			<div class="empty-state">
				<Compass size={48} />
				<p>No content found for this view.</p>
				<button class="btn-retry" onclick={refresh}>Try Again</button>
			</div>
		{:else}
			<div class={clsx('content-layout', layout)}>
				{#each content as item (item.url)}
					<DocumentGridItem
						title={item.title}
						cover_url={item.cover_url}
						meta={item.info}
						onclick={() => onSelectDocument(item.url, pluginName)}
					/>
				{/each}
			</div>
		{/if}
	</div>
</BasePage>

<style>
	.actions-group {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: var(--bg-surface);
		padding: 0.25rem;
		border-radius: 10px;
		border: 1px solid var(--border-main);
	}

	.icon-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		padding: 0.5rem;
		cursor: pointer;
		border-radius: 8px;
		display: flex;
		align-items: center;
		transition: all 0.2s;
	}

	.icon-btn:hover {
		color: var(--text-main);
		background: rgba(255, 255, 255, 0.05);
	}
	.icon-btn.active {
		color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}
	.icon-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.divider {
		width: 1px;
		height: 20px;
		background: var(--border-main);
		margin: 0 0.25rem;
	}

	.explorer-content {
		width: 100%;
	}

	.content-layout.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 2rem;
	}

	.loading-state,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8rem 0;
		color: var(--text-muted);
		gap: 1.5rem;
	}

	.spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--bg-surface);
		border-top-color: var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.btn-retry {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 10px;
		font-weight: 700;
		cursor: pointer;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}
	@keyframes animate-spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}
	.animate-spin {
		animation: animate-spin 1s linear infinite;
	}

	.capitalize {
		text-transform: capitalize;
	}
</style>
