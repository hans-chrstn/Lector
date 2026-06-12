<script lang="ts">
	import SearchIcon from 'lucide-svelte/icons/search';
	import Loader2 from 'lucide-svelte/icons/loader-2';
	import ExternalLink from 'lucide-svelte/icons/external-link';
	import DocumentGridItem from '../components/DocumentGridItem.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { toast } from '../services/toast.svelte';
	import type { SearchItem, PluginManifest } from '$lib/services/api';

	interface Props {
		plugins: PluginManifest[];
		results: SearchItem[];
		loading: boolean;
		query?: string;
		source?: string;
		onSearch: (query: string, source: string) => void;
		onSelect: (url: string, source: string) => void;
	}
	let {
		plugins,
		results,
		loading,
		query = $bindable(''),
		source = $bindable('library'),
		onSearch,
		onSelect
	}: Props = $props();

	const filteredSources = $derived(
		plugins.filter((p) => p.is_enabled && p.capabilities?.includes('catalog')).map((p) => p.name)
	);

	let timeoutId: any;

	$effect(() => {
		const q = query;
		const s = source;
		if (timeoutId) {
			clearTimeout(timeoutId);
		}
		if (q.trim() === '') {
			return;
		}
		timeoutId = setTimeout(() => {
			onSearch(q, s);
		}, 300);
		return () => {
			if (timeoutId) {
				clearTimeout(timeoutId);
			}
		};
	});

	function triggerSearch() {
		if (timeoutId) clearTimeout(timeoutId);
		onSearch(query, source);
	}

	const searchActions = $derived.by(() => {
		const actionMap = new SvelteMap<string, any>();
		if (source === 'all') return [];
		plugins.forEach((p) => {
			if (!p.is_enabled) return;
			if (source !== p.name) return;
			(p.actions || [])
				.filter((a) => a.context === 'search')
				.forEach((a) => actionMap.set(a.label, { ...a, plugin: p.name }));
		});
		return Array.from(actionMap.values());
	});

	let executingAction = $state(false);
	async function handleActionClick(action: any) {
		if (executingAction) return;
		executingAction = true;
		try {
			const res = await fetch(
				`${window.location.origin}/api/plugins/${action.plugin}/rpc/${action.method}`,
				{
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ query })
				}
			);
			if (!res.ok) {
				toast.error(await res.text());
			} else {
				const data = await res.json();
				if (data.open_url) {
					window.open(data.open_url, '_blank');
				} else if (data.message) {
					toast.success(data.message);
				}
			}
		} catch (e) {
			console.error(`[Search] Failed to execute action ${action.method}`, e);
			toast.error(`Action failed: ${e}`);
		} finally {
			executingAction = false;
		}
	}
</script>

<BasePage title="Discovery" subtitle="Find new titles across all your enabled sources">
	{#snippet actions()}
		<div class="search-bar">
			<div class="input-group">
				<SearchIcon size={20} class="search-icon" />
				<input
					type="text"
					bind:value={query}
					placeholder="Search titles..."
					onkeydown={(e) => e.key === 'Enter' && triggerSearch()}
				/>
				<div class="divider"></div>
				<select bind:value={source} class="source-select">
					<option value="library">Library</option>
					<option value="all">All Sources</option>
					{#each filteredSources as s (s)}
						<option value={s}>{s.replace(/-/g, ' ')}</option>
					{/each}
				</select>
			</div>
			<button class="search-btn" onclick={triggerSearch} disabled={loading}>
				{#if loading}
					<Loader2 size={18} class="spin" />
				{:else}
					<span>Search</span>
				{/if}
			</button>
		</div>
	{/snippet}

	{#if searchActions.length > 0}
		<div class="plugin-actions-row">
			{#each searchActions as action (action.label)}
				<button class="action-chip" onclick={() => handleActionClick(action)}>
					<ExternalLink size={14} />
					<span>{action.label}</span>
				</button>
			{/each}
		</div>
	{/if}

	<div class="grid">
		{#each results as res, i (res.source + '_' + res.url + '_' + i)}
			<DocumentGridItem
				title={res.title}
				cover_url={res.cover_url}
				meta={res.info}
				onclick={() =>
					onSelect(
						res.url,
						source === 'all' || source === 'library' ? (res as any).source : source
					)}
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
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 0 1rem;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.input-group:focus-within {
		border-color: var(--primary);
		background: rgba(255, 255, 255, 0.06);
		box-shadow: 0 0 0 2px rgba(var(--primary-rgb), 0.1);
	}

	:global(.search-icon) {
		color: var(--text-muted);
		min-width: 20px;
	}

	input {
		background: none;
		border: none;
		color: var(--text-main);
		height: 44px;
		width: 100%;
		font-size: 0.9rem;
		font-weight: 500;
		outline: none;
	}

	.search-btn {
		background: white;
		color: black;
		border: none;
		padding: 0 1.5rem;
		border-radius: 12px;
		font-weight: 700;
		font-size: 0.875rem;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		display: flex;
		align-items: center;
		justify-content: center;
		min-width: 100px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.search-btn:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 0 20px rgba(255, 255, 255, 0.2);
		background: #f4f4f5;
	}

	.search-btn:active:not(:disabled) {
		transform: translateY(0);
	}

	.search-btn:disabled {
		opacity: 0.5;
		background: var(--bg-surface-hover);
		color: var(--text-muted);
		cursor: not-allowed;
		box-shadow: none;
	}

	.divider {
		width: 1px;
		height: 24px;
		background: var(--border-main);
		margin: 0 0.5rem;
	}

	.source-select {
		background: transparent;
		border: none;
		color: var(--text-main);
		font-size: 0.875rem;
		font-weight: 600;
		outline: none;
		cursor: pointer;
		text-transform: capitalize;
		padding-right: 0.5rem;
	}

	.source-select option {
		background: var(--bg-main);
		color: var(--text-main);
	}

	.plugin-actions-row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-bottom: 2rem;
	}

	.action-chip {
		background: rgba(var(--primary-rgb), 0.1);
		border: 1px solid rgba(var(--primary-rgb), 0.2);
		color: var(--primary);
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

	.action-chip:hover {
		background: rgba(var(--primary-rgb), 0.2);
		border-color: rgba(var(--primary-rgb), 0.3);
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

	:global(.spin) {
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
		.input-group {
			flex-direction: column;
			align-items: stretch;
			background: transparent;
			border: none;
			padding: 0;
		}
		.input-group input,
		.input-group select {
			background: rgba(255, 255, 255, 0.03);
			border: 1px solid var(--border-main);
			padding: 0 1rem;
			border-radius: 12px;
			width: 100%;
			height: 44px;
		}
		.divider {
			display: none;
		}
	}
</style>
