<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import {
		Loader2,
		AlertCircle,
		GripVertical,
		Trash2,
		Download,
		DownloadCloud,
		Zap,
		Settings,
		Info,
		CheckCircle2,
		History,
		Database,
		Layout,
		Globe,
		ShieldCheck
	} from 'lucide-svelte';
	import { dndzone } from 'svelte-dnd-action';
	import BasePage from '../components/base/BasePage.svelte';
	import { clsx } from 'clsx';

	interface Props {
		pluginName: string;
		tabId: string;
	}
	let { pluginName, tabId }: Props = $props();

	let schema = $state<any>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let interval: any;

	const iconMap: Record<string, any> = {
		Trash2,
		Download,
		DownloadCloud,
		Zap,
		Settings,
		Info,
		CheckCircle2,
		History,
		Database,
		Layout,
		Globe,
		ShieldCheck
	};

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

	async function handleReorder(componentId: string, items: any[]) {
		try {
			await fetch(`${getBase()}/api/plugins/${pluginName}/rpc/on_reorder`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ component_id: componentId, items })
			});
		} catch (e) {
			console.error('Failed to notify plugin of reorder:', e);
		}
	}

	async function handlePluginAction(pluginName: string, method: string, args: any) {
		try {
			const res = await fetch(`${getBase()}/api/plugins/${pluginName}/rpc/${method}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(args)
			});
			if (!res.ok) alert(await res.text());
			else {
				const data = await res.json();
				if (data.message) alert(data.message);
				fetchSchema(false);
			}
		} catch (e) {
			console.error('Failed to execute plugin action:', e);
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
				{:else if comp.type === 'SortableList'}
					<div
						class="sortable-list"
						use:dndzone={{ items: comp.props.items, flipDurationMs: 200, type: comp.id }}
						onconsider={(e) => {
							const idx = schema.components.findIndex((c: any) => c.id === comp.id);
							schema.components[idx].props.items = e.detail.items;
						}}
						onfinalize={(e) => {
							const idx = schema.components.findIndex((c: any) => c.id === comp.id);
							schema.components[idx].props.items = e.detail.items;
							handleReorder(comp.id, e.detail.items);
						}}
					>
						{#each comp.props.items || [] as item (item.id)}
							<div class="sortable-item">
								<div class="item-main">
									<div class="grab-handle">
										<GripVertical size={16} />
									</div>
									<div class="item-content">
										<div class="content-left">
											<span class="title">{item.title}</span>
											{#if item.subtitle}
												<span class="subtitle">{item.subtitle}</span>
											{/if}
										</div>
										<div class="content-right">
											<div class="status-group">
												{#if item.download_url}
													<a href={item.download_url} class="download-link" download rel="external"
														>Download</a
													>
												{/if}
												{#if item.status}
													<span class={clsx('status-badge', item.status_variant || 'neutral')}>
														{item.status}
													</span>
												{/if}
											</div>
											{#if item.actions}
												<div class="item-actions">
													{#each item.actions as action (action.method)}
														<button
															class="action-btn"
															onclick={() =>
																handlePluginAction(pluginName, action.method, { id: item.id })}
															title={action.label}
														>
															{#if action.icon && iconMap[action.icon]}
																{@const Icon = iconMap[action.icon]}
																<Icon size={14} />
															{:else}
																<Zap size={14} />
															{/if}
														</button>
													{/each}
												</div>
											{/if}
										</div>
									</div>
								</div>
								{#if item.progress !== undefined}
									<div class="item-progress">
										<div class="track">
											<div class="fill" style="width: {item.progress}%"></div>
										</div>
									</div>
								{/if}
							</div>
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

	.progress-list,
	.sortable-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.progress-item,
	.sortable-item {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 1.25rem;
	}
	.sortable-item {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}
	.item-main {
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	.grab-handle {
		color: var(--text-dim);
		cursor: grab;
	}
	.item-content {
		flex: 1;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.content-left {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.subtitle {
		font-size: 0.75rem;
		color: var(--text-dim);
		font-weight: 400;
	}
	.content-right {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}
	.status-badge {
		padding: 0.25rem 0.6rem;
		border-radius: 6px;
		font-size: 0.7rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.status-badge.neutral {
		background: var(--bg-surface-hover);
		color: var(--text-dim);
	}
	.status-badge.success {
		background: #10b98120;
		color: #10b981;
	}
	.status-badge.warning {
		background: #f59e0b20;
		color: #f59e0b;
	}
	.status-badge.error {
		background: #ef444420;
		color: #ef4444;
	}

	.item-actions {
		display: flex;
		gap: 0.5rem;
	}
	.action-btn {
		background: var(--bg-surface-hover);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.4rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
	}
	.action-btn:hover {
		border-color: var(--primary);
		color: var(--primary);
	}

	.item-progress {
		width: 100%;
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
		background: var(--text-main);
		color: var(--bg-main);
		padding: 0.3rem 0.8rem;
		border-radius: 8px;
		text-decoration: none;
		font-weight: 700;
		font-size: 0.75rem;
		transition: all 0.2s;
	}
	.download-link:hover {
		opacity: 0.9;
		transform: translateY(-1px);
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
