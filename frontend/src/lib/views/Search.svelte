<script lang="ts">
	import { Search as SearchIcon, Loader2, Globe, Compass } from 'lucide-svelte';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import { clsx } from 'clsx';
	import type { SearchItem } from '$lib/services/api';

	interface Props {
		sources: string[];
		results: SearchItem[];
		loading: boolean;
		onSearch: (query: string, source: string) => void;
		onSelect: (url: string, source: string) => void;
	}
	let { sources, results, loading, onSearch, onSelect }: Props = $props();

	let query = $state('');
	let source = $state('all');

	const filteredSources = $derived(
		sources.filter((s) => s !== 'system_manager' && s !== 'source_manager')
	);
</script>

<BasePage title="Discovery" subtitle="Find new titles across all your enabled sources">
	{#snippet actions()}
		<div class="search-bar">
			<div class="input-group">
				<SearchIcon size={18} />
				<input
					type="text"
					bind:value={query}
					placeholder="Search titles..."
					onkeydown={(e) => e.key === 'Enter' && onSearch(query, source)}
				/>
			</div>
			<button class="search-btn" onclick={() => onSearch(query, source)} disabled={loading}>
				{#if loading}
					<Loader2 size={18} class="spin" />
				{:else}
					<span>Search</span>
				{/if}
			</button>
		</div>
	{/snippet}

	{#snippet extraHeader()}
		<div class="source-picker">
			<button
				class={clsx('source-chip', source === 'all' && 'active')}
				onclick={() => (source = 'all')}
			>
				<Globe size={14} />
				<span>All Sources</span>
			</button>
			{#each filteredSources as s (s)}
				<button class={clsx('source-chip', source === s && 'active')} onclick={() => (source = s)}>
					<Compass size={14} />
					<span class="capitalize">{s}</span>
				</button>
			{/each}
		</div>
	{/snippet}

	<div class="grid">
		{#each results as res (res.url)}
			<DocumentGridItem
				title={res.title}
				cover_url={res.cover_url}
				meta={res.info}
				onclick={() => onSelect(res.url, source === 'all' ? (res as any).source : source)}
			/>
		{:else}
			{#if !loading}
				<div class="empty-results">
					<SearchIcon size={48} strokeWidth={1} />
					<p>Enter a keyword to start discovery</p>
				</div>
			{/if}
		{/each}
	</div>
</BasePage>

<style>
	.search-bar {
		display: flex;
		gap: 0.5rem;
		width: 400px;
	}

	.input-group {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 10px;
		padding: 0 1rem;
		transition: all 0.2s;
	}

	.input-group:focus-within {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
	}

	input {
		background: none;
		border: none;
		color: var(--text-main);
		height: 40px;
		width: 100%;
		font-size: 0.875rem;
		outline: none;
	}

	.search-btn {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0 1.25rem;
		border-radius: 10px;
		font-weight: 700;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.search-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.source-picker {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-bottom: 2rem;
		border-bottom: 1px solid var(--border-main);
		padding-bottom: 1.5rem;
	}

	.source-chip {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-dim);
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.8125rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
		cursor: pointer;
	}

	.source-chip:hover {
		border-color: var(--primary);
		color: var(--text-main);
	}

	.source-chip.active {
		background: var(--primary);
		color: white;
		border-color: var(--primary);
	}

	.capitalize {
		text-transform: capitalize;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 2rem;
	}

	.empty-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8rem 0;
		color: var(--text-dim);
		text-align: center;
		gap: 1rem;
	}

	.spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (max-width: 900px) {
		.search-bar {
			width: 100%;
		}
	}
</style>
