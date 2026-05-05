<script lang="ts">
	import { Clock, Ghost } from 'lucide-svelte';
	import DocumentListItem from '../components/DocumentListItem.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	import type { Document } from '$lib/services/api';

	interface Props {
		history: Document[];
		onOpenDocument: (document: Document) => void;
	}
	let { history, onOpenDocument }: Props = $props();
</script>

<BasePage title="History" subtitle="Pick up where you left off">
	{#snippet actions()}
		<div class="status-badge active">
			<Clock size={14} />
			<span>Recently Read</span>
		</div>
	{/snippet}

	<div class="list-container">
		{#each history as document (document.id)}
			<DocumentListItem
				title={document.title}
				cover_url={document.cover_url}
				meta={document.source}
				is_local={document.is_local}
				read_chapters={document.read_chapters}
				total_chapters={document.chapters?.length || 0}
				onclick={() => onOpenDocument(document)}
			/>
		{:else}
			<div class="empty-state">
				<Ghost size={40} class="empty-icon" />
				<h3>No history yet</h3>
				<p>Start reading to see your recently visited titles here.</p>
			</div>
		{/each}
	</div>
</BasePage>

<style>
	.status-badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.8125rem;
		font-weight: 600;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
	}

	.status-badge.active {
		color: var(--primary);
		border-color: rgba(59, 130, 246, 0.2);
		background: rgba(59, 130, 246, 0.05);
	}

	.list-container {
		display: flex;
		flex-direction: column;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 8rem 0;
		text-align: center;
	}

	.empty-state h3 {
		font-size: 1.125rem;
		font-weight: 600;
		margin: 0 0 0.5rem;
	}
	.empty-state p {
		color: var(--text-dim);
		font-size: 0.875rem;
		max-width: 300px;
		margin: 0;
	}
</style>
