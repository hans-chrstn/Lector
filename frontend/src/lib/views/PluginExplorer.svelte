<script lang="ts">
	import Loader2 from 'lucide-svelte/icons/loader-2';
	import Info from 'lucide-svelte/icons/info';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import { api, type SearchItem } from '$lib/services/api';
	import { clsx } from 'clsx';

	interface Props {
		name: string;
		popular: SearchItem[];
		latest: SearchItem[];
		loading: boolean;
		onSelect: (url: string, source: string) => void;
	}
	let { name, popular = [], latest = [], loading, onSelect }: Props = $props();

	let tab = $state('popular');
	let items = $state<SearchItem[]>([]);
	let page = $state(1);
	let infiniteLoading = $state(false);
	let hasMore = $state(true);

	$effect(() => {
		items = tab === 'popular' ? popular : latest;
		page = 1;
		hasMore = true;
	});

	async function loadMore() {
		if (infiniteLoading || !hasMore) return;
		infiniteLoading = true;
		try {
			const next =
				tab === 'popular'
					? await api.getDocumentPopular(name, page + 1)
					: await api.getDocumentLatest(name, page + 1);

			if (next.length === 0) hasMore = false;
			else {
				items = [...items, ...next];
				page++;
			}
		} finally {
			infiniteLoading = false;
		}
	}
</script>

<BasePage
	title={name}
	subtitle="Discover trending and new titles from this source"
	containerClass="capitalize"
>
	{#snippet actions()}
		<div class="tabs">
			<button class={clsx('tab', tab === 'popular' && 'active')} onclick={() => (tab = 'popular')}>
				Popular
			</button>
			<button class={clsx('tab', tab === 'latest' && 'active')} onclick={() => (tab = 'latest')}>
				Latest
			</button>
		</div>
	{/snippet}

	{#if loading && items.length === 0}
		<div class="loading-state">
			<Loader2 size={32} class="spin" />
			<p>Syncing plugin index</p>
		</div>
	{:else}
		<div class="grid">
			{#each items as res (res.url)}
				<DocumentGridItem
					title={res.title}
					cover_url={res.cover_url}
					onclick={() => onSelect(res.url, name)}
				/>
			{:else}
				<div class="empty-state">
					<Info size={40} />
					<h3>No results found</h3>
					<p>Try refreshing the plugin or checking your connection.</p>
				</div>
			{/each}
		</div>

		{#if hasMore && items.length > 0}
			<div class="footer-actions">
				<button class="load-more" onclick={loadMore} disabled={infiniteLoading}>
					{#if infiniteLoading}
						<Loader2 size={18} class="spin" />
						<span>Loading more...</span>
					{:else}
						<span>Load More</span>
					{/if}
				</button>
			</div>
		{/if}
	{/if}
</BasePage>

<style>
	.tabs {
		display: flex;
		background: var(--bg-surface);
		padding: 2px;
		border-radius: 10px;
		border: 1px solid var(--border-main);
	}

	.tab {
		padding: 0.5rem 1.25rem;
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 700;
		color: var(--text-dim);
		cursor: pointer;
		border: none;
		background: none;
		transition: all 0.2s;
	}

	.tab:hover {
		color: var(--text-main);
	}

	.tab.active {
		background: var(--bg-main);
		color: var(--primary);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
	}

	.grid {
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
		color: var(--text-dim);
		gap: 1.5rem;
		text-align: center;
	}

	.empty-state h3 {
		color: var(--text-main);
		margin: 0;
	}

	.footer-actions {
		display: flex;
		justify-content: center;
		margin-top: 4rem;
	}

	.load-more {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.75rem 2.5rem;
		border-radius: 12px;
		font-weight: 700;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		transition: all 0.2s;
	}

	.load-more:hover:not(:disabled) {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
		transform: translateY(-2px);
	}

	.spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.capitalize {
		text-transform: capitalize;
	}
</style>
