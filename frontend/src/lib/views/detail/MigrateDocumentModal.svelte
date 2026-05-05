<script lang="ts">
	import { Loader2 } from 'lucide-svelte';
	import Modal from '../../components/Modal.svelte';
	import { api, type SearchItem } from '$lib/services/api';

	interface Props {
		show: boolean;
		sources: string[];
		onClose: () => void;
		onSelect: (result: SearchItem, source: string) => void;
	}

	let { show, sources, onClose, onSelect }: Props = $props();

	let migrateSource = $state('');
	let migrateQuery = $state('');
	let migrateResults = $state<SearchItem[]>([]);
	let isMigrating = $state(false);

	$effect(() => {
		if (show && sources.length > 0 && !migrateSource) {
			migrateSource = sources.filter((s) => s !== 'local')[0] || '';
		}
	});

	async function handleMigrateSearch() {
		if (!migrateQuery.trim()) return;
		isMigrating = true;
		try {
			migrateResults = await api.search(migrateSource, migrateQuery);
		} finally {
			isMigrating = false;
		}
	}
</script>

<Modal {show} title="Migrate Document" {onClose} width="600px">
	<div class="migrate-modal">
		<p class="desc">Switch this document to a different scraper while keeping your history.</p>
		<div class="search-box">
			<select bind:value={migrateSource}>
				{#each sources.filter((s) => s !== 'local') as s (s)}
					<option value={s}>{s}</option>
				{/each}
			</select>
			<input
				type="text"
				bind:value={migrateQuery}
				placeholder="Search target..."
				onkeydown={(e) => e.key === 'Enter' && handleMigrateSearch()}
			/>
			<button class="primary-btn-small" onclick={handleMigrateSearch} disabled={isMigrating}>
				{#if isMigrating}<Loader2 size={16} class="spin" />{:else}Search{/if}
			</button>
		</div>
		<div class="results-list">
			{#each migrateResults as res (res.url)}
				<button type="button" class="result-item" onclick={() => onSelect(res, migrateSource)}>
					<img src={api.getProxyImage(res.cover_url)} alt="" class="mini-cover" />
					<div class="res-info">
						<span class="res-title">{res.title}</span>
						<span class="res-meta">{res.info}</span>
					</div>
				</button>
			{:else}
				<div class="empty-results">Search for a replacement source</div>
			{/each}
		</div>
	</div>
</Modal>

<style>
	.migrate-modal .desc {
		color: var(--text-dim);
		font-size: 0.9rem;
		margin-bottom: 1.5rem;
	}
	.search-box {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
	}
	select,
	input {
		background: var(--bg-main);
		border: 1px solid var(--border-bright);
		color: var(--text-main);
		padding: 0.75rem;
		border-radius: 8px;
		font-size: 0.9rem;
		outline: none;
	}
	.primary-btn-small {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0 1.25rem;
		border-radius: 8px;
		font-weight: 700;
		cursor: pointer;
	}
	.results-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		max-height: 400px;
		overflow-y: auto;
	}
	.result-item {
		display: flex;
		gap: 1rem;
		padding: 0.75rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		cursor: pointer;
		text-align: left;
		width: 100%;
		transition: all 0.2s;
	}
	.result-item:hover {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
	}
	.mini-cover {
		width: 50px;
		aspect-ratio: 2/3;
		border-radius: 6px;
		object-fit: cover;
		background: var(--bg-main);
	}
	.res-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.res-title {
		font-weight: 700;
		font-size: 0.9rem;
		color: var(--text-main);
	}
	.res-meta {
		font-size: 0.75rem;
		color: var(--text-dim);
	}
	.empty-results {
		text-align: center;
		padding: 2rem;
		color: var(--text-dim);
		font-size: 0.875rem;
	}
</style>
