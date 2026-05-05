<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Loader2, AlertCircle } from 'lucide-svelte';
	import BasePage from '../components/base/BasePage.svelte';

	interface Props {
		pluginName: string;
		tabId: string;
	}
	let { pluginName, tabId }: Props = $props();

	let schema = $state<any>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let interval: any;

	const getBase = () => {
		if (typeof window !== 'undefined') return window.location.origin;
		return 'http://localhost:3000';
	};

	async function fetchSchema(showLoading = true) {
		if (showLoading) loading = true;
		error = null;
		try {
			const res = await fetch(`${getBase()}/api/plugins/${pluginName}/rpc/get_ui_schema`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ tab_id: tabId })
			});
			if (!res.ok) throw new Error(await res.text());
			schema = await res.json();
		} catch (err: any) {
			error = err.message || 'Failed to load plugin UI';
		} finally {
			if (showLoading) loading = false;
		}
	}

	onMount(() => {
		fetchSchema(true);
		interval = setInterval(() => fetchSchema(false), 2000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});

	$effect(() => {
		if (pluginName && tabId) {
			fetchSchema(true);
		}
	});
</script>

{#if loading && !schema}
	<div class="center-state">
		<Loader2 size={32} class="spin text-primary" />
		<p>Loading plugin interface...</p>
	</div>
{:else if error && !schema}
	<div class="center-state text-error">
		<AlertCircle size={32} />
		<p>{error}</p>
		<button class="retry-btn" onclick={() => fetchSchema(true)}>Retry</button>
	</div>
{:else if schema}
	<BasePage
		title={schema.title || pluginName}
		subtitle={schema.subtitle || `Managed by ${pluginName}`}
		containerClass="capitalize"
	>
		<div class={`layout-${schema.layout || 'vertical'}`}>
			{#each schema.components || [] as comp, i (i)}
				{#if comp.type === 'Header'}
					<div class="dynamic-section-header">
						<h2>{comp.props.title}</h2>
						{#if comp.props.subtitle}
							<p>{comp.props.subtitle}</p>
						{/if}
					</div>
				{:else if comp.type === 'ProgressList'}
					<div class="progress-list">
						{#each comp.props.items || [] as item, j (j)}
							<div class="progress-item">
								<div class="info">
									<span class="title">{item.title}</span>
									<div class="status-group">
										{#if item.download_url}
											<a href={item.download_url} class="download-link" download rel="external"
												>Download</a
											>
										{/if}
										<span class="status">{item.status}</span>
									</div>
								</div>
								<div class="track">
									<div class="fill" style="width: {item.progress}%"></div>
								</div>
							</div>
						{:else}
							<p class="empty-text">No active items.</p>
						{/each}
					</div>
				{:else}
					<div class="unknown-component">Unsupported component type: {comp.type}</div>
				{/if}
			{/each}
		</div>
	</BasePage>
{/if}

<style>
	.center-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 50vh;
		gap: 1rem;
		color: var(--text-dim);
	}
	.text-primary {
		color: var(--primary);
	}
	.text-error {
		color: #ef4444;
	}
	.retry-btn {
		background: none;
		border: 1px solid currentColor;
		color: inherit;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
	}
	.spin {
		animation: spin 1s linear infinite;
	}

	.layout-vertical {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.dynamic-section-header h2 {
		font-size: 1.25rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
	}
	.dynamic-section-header p {
		color: var(--text-dim);
		margin: 0;
		font-size: 0.875rem;
	}

	.progress-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.progress-item {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 1.25rem;
	}
	.progress-item .info {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1rem;
		font-size: 0.9rem;
		font-weight: 600;
	}
	.progress-item .status {
		color: var(--text-dim);
		text-transform: capitalize;
		font-size: 0.8rem;
	}
	.status-group {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.download-link {
		color: var(--primary);
		text-decoration: underline;
		font-weight: 700;
		font-size: 0.8rem;
	}
	.track {
		height: 6px;
		background: var(--bg-main);
		border-radius: 10px;
		overflow: hidden;
	}
	.fill {
		height: 100%;
		background: var(--primary);
		transition: width 0.3s;
	}
	.empty-text {
		color: var(--text-dim);
		text-align: center;
		padding: 2rem;
		font-style: italic;
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
