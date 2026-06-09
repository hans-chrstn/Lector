<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import Loader2 from 'lucide-svelte/icons/loader-2';
	import AlertCircle from 'lucide-svelte/icons/alert-circle';
	import Zap from 'lucide-svelte/icons/zap';
	import Settings from 'lucide-svelte/icons/settings';
	import Info from 'lucide-svelte/icons/info';
	import CheckCircle2 from 'lucide-svelte/icons/check-circle-2';
	import History from 'lucide-svelte/icons/history';
	import Database from 'lucide-svelte/icons/database';
	import Layout from 'lucide-svelte/icons/layout';
	import Globe from 'lucide-svelte/icons/globe';
	import ShieldCheck from 'lucide-svelte/icons/shield-check';
	import X from 'lucide-svelte/icons/x';
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import Download from 'lucide-svelte/icons/download';
	import DownloadCloud from 'lucide-svelte/icons/download-cloud';
	import GripVertical from 'lucide-svelte/icons/grip-vertical';
	import { dndzone } from 'svelte-dnd-action';
	import BasePage from '../components/base/BasePage.svelte';
	import { toast } from '$lib/services/toast.svelte';
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
	let formData = $state<Record<string, string>>({});

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
		ShieldCheck,
		X
	};

	onMount(() => {
		window.addEventListener('message', handleMessage);

		// Ping IFrames that we are ready to receive their initial data
		const iframes = document.querySelectorAll('iframe');
		iframes.forEach((iframe) => {
			if (iframe.contentWindow) {
				iframe.contentWindow.postMessage({ type: 'READY' }, '*');
			}
		});

		return () => {
			window.removeEventListener('message', handleMessage);
		};
	});

	function handleMessage(e: MessageEvent) {
		if (e.data?.type === 'UPDATE_FORM_DATA' && e.data?.id) {
			formData[e.data.id] = e.data.value;
		}
	}

	$effect(() => {
		if (pluginName && tabId) {
			schema = null;
			fetchSchema(true);
		}
	});

	// Re-ping iframes when schema updates and renders them
	$effect(() => {
		if (schema && !loading) {
			setTimeout(() => {
				const iframes = document.querySelectorAll('iframe');
				iframes.forEach((iframe) => {
					if (iframe.contentWindow) {
						iframe.contentWindow.postMessage({ type: 'READY' }, '*');
					}
				});
			}, 200); // Give DOM time to render iframe tags
		}
	});

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

			const contentType = res.headers.get('content-type');
			if (!res.ok) {
				const text = await res.text();
				toast.error(text || `Action failed (${res.status})`);
				return;
			}

			if (contentType && contentType.includes('application/json')) {
				const data = await res.json();
				if (data.download_url) {
					const link = window.document.createElement('a');
					link.href = data.download_url;
					link.download = '';
					window.document.body.appendChild(link);
					link.click();
					window.document.body.removeChild(link);
				}
				if (data.message) toast.success(data.message);
				fetchSchema(false);
			} else {
				toast.success('Action executed');
				fetchSchema(false);
			}
		} catch (e: any) {
			console.error('Plugin action error:', e);
			toast.error(e.name === 'TypeError' ? 'Connection lost or blocked' : 'Action failed');
		}
	}

	onMount(() => {
		interval = setInterval(() => fetchSchema(false), 5000);
	});

	onDestroy(() => {
		if (interval) clearInterval(interval);
	});
</script>

<BasePage
	title={schema?.title || pluginName.charAt(0).toUpperCase() + pluginName.slice(1)}
	subtitle={schema?.subtitle || ''}
>
	<div class="dynamic-container">
		{#if loading && !schema}
			<div class="loader-box">
				<Loader2 size={32} class="spin" />
				<p>Loading plugin interface...</p>
			</div>
		{:else if error}
			<div class="error-box">
				<AlertCircle size={32} />
				<h3>Interface Error</h3>
				<p>{error}</p>
				<button class="retry-btn" onclick={() => fetchSchema(true)}>Retry Connection</button>
			</div>
		{:else if schema}
			{@const components = Array.isArray(schema) ? schema : schema.components || []}
			{#if Array.isArray(components) && components.length > 0}
				<div class={schema.layout === 'grid' ? 'components-grid' : 'components-linear'}>
					{#each components as component (component.id)}
						<div
							class="component-wrapper"
							style={schema.layout === 'grid' && component.type === 'Button'
								? 'grid-column: 1 / -1;'
								: ''}
						>
							{#if component.type === 'Text'}
								<div class="text-block">
									{#if component.props?.title}<h4>{component.props.title}</h4>{/if}
									<p>{component.props?.text}</p>
								</div>
							{:else if component.type === 'Header'}
								<div class="header-block">
									<h4>{component.props?.title}</h4>
									<p>{component.props?.subtitle}</p>
								</div>
							{:else if component.type === 'SortableList'}
								<div class="list-section">
									<header>
										<div class="title-info">
											<h4>{component.props?.title}</h4>
											<p>{component.props?.subtitle}</p>
										</div>
									</header>
									{#if component.props?.items && Array.isArray(component.props.items)}
										<div
											class="dynamic-list"
											use:dndzone={{ items: component.props.items, flipDurationMs: 200 }}
											onconsider={(e) => (component.props.items = e.detail.items)}
											onfinalize={(e) => handleReorder(component.id, e.detail.items)}
										>
											{#each component.props.items as item (item.id)}
												<div class="list-item-card">
													<div class="drag-handle"><GripVertical size={16} /></div>
													{#if item.cover_url}
														<img src={item.cover_url} alt="" class="item-cover" />
													{/if}
													<div class="item-meta">
														<div class="title-row">
															<span class="item-title">{item.title}</span>
															{#if item.status}
																<span
																	class={clsx('status-badge', item.status_variant || item.status)}
																>
																	{item.status}
																</span>
															{/if}
														</div>
														<span class="item-sub">{item.subtitle}</span>
														{#if item.progress !== undefined}
															<div class="item-progress">
																<div class="bar" style="width: {item.progress}%"></div>
															</div>
														{/if}
													</div>
													<div class="item-actions">
														{#each item.actions || [] as action (action.label)}
															{@const Icon = iconMap[action.icon] || Zap}
															<button
																class="action-icon-btn"
																onclick={() => handlePluginAction(pluginName, action.method, item)}
																title={action.label}
															>
																<Icon size={16} />
															</button>
														{/each}
													</div>
												</div>
											{/each}
										</div>
									{/if}
								</div>
							{:else if component.type === 'TextInput'}
								<div class="input-field">
									{#if component.props?.label}
										<label for={component.id}>{component.props.label}</label>
									{/if}
									<input
										id={component.id}
										type={component.props?.type || 'text'}
										placeholder={component.props?.placeholder || ''}
										bind:value={formData[component.id]}
										class={component.props?.type === 'color' ? 'color-input' : ''}
									/>
								</div>
							{:else if component.type === 'IFrame'}
								<div class="iframe-field">
									{#if component.props?.label}
										<span class="field-label">{component.props.label}</span>
									{/if}
									<iframe
										src={component.props?.src}
										style={component.props?.style ||
											'width: 100%; height: 250px; border: none; border-radius: 8px; background: transparent;'}
										title={component.props?.label || component.id}
									></iframe>
								</div>
							{:else if component.type === 'Button'}
								<button
									class="primary-btn"
									onclick={() => handlePluginAction(pluginName, component.props.method, formData)}
								>
									{component.props?.label || 'Submit'}
								</button>
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<div class="empty-state">
					<Zap size={48} />
					<p>This plugin has no active interface elements</p>
				</div>
			{/if}
		{/if}
	</div>
</BasePage>

<style>
	.dynamic-container {
		display: flex;
		flex-direction: column;
		gap: 2rem;
		max-width: 900px;
		margin: 0 auto;
	}

	.loader-box,
	.error-box,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 6rem 2rem;
		color: var(--text-dim);
		text-align: center;
	}

	.error-box {
		color: #ef4444;
		background: rgba(239, 68, 68, 0.05);
		border: 1px solid rgba(239, 68, 68, 0.1);
		border-radius: 20px;
	}

	.error-box h3 {
		margin: 1rem 0 0.5rem;
		color: var(--text-main);
	}

	.retry-btn {
		margin-top: 1.5rem;
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0.6rem 1.5rem;
		border-radius: 10px;
		font-weight: 700;
		cursor: pointer;
	}

	.components-linear {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.components-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
		align-items: start;
	}

	.header-block {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid var(--border-main);
	}

	.header-block h4 {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--text-main);
	}

	.header-block p {
		margin: 0.5rem 0 0;
		color: var(--text-dim);
	}

	.text-block h4 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
	}

	.text-block p {
		color: var(--text-dim);
		line-height: 1.6;
	}

	.list-section header {
		margin-bottom: 1.25rem;
	}

	.title-info h4 {
		margin: 0;
		font-size: 1.2rem;
		font-weight: 800;
	}

	.title-info p {
		margin: 0.25rem 0 0;
		font-size: 0.85rem;
		color: var(--text-dim);
	}

	.dynamic-list {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.list-item-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 0.75rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.2s;
	}

	.color-picker-container {
		display: flex;
		justify-content: center;
		padding: 1rem;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		border-radius: 8px;
		margin-top: 0.5rem;
	}

	.color-input {
		-webkit-appearance: none;
		-moz-appearance: none;
		appearance: none;
		width: 48px;
		height: 48px;
		border: none;
		border-radius: 50%;
		padding: 0;
		cursor: pointer;
		overflow: hidden;
		background: none;
	}
	.color-input::-webkit-color-swatch-wrapper {
		padding: 0;
	}
	.color-input::-webkit-color-swatch {
		border: 2px solid var(--border-main);
		border-radius: 50%;
	}
	.color-input::-moz-color-swatch {
		border: 2px solid var(--border-main);
		border-radius: 50%;
	}

	.primary-btn {
		width: 100%;
		padding: 0.85rem;
		background: var(--primary);
		color: #ffffff;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.95rem;
		cursor: pointer;
		transition: opacity 0.2s;
		margin-top: 1rem;
	}

	.list-item-card:hover {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
	}

	.drag-handle {
		cursor: grab;
		color: var(--text-dim);
		padding: 0.5rem;
	}

	.item-cover {
		width: 44px;
		height: 60px;
		object-fit: cover;
		border-radius: 6px;
		background: var(--bg-main);
	}

	.item-meta {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.item-title {
		font-weight: 700;
		font-size: 0.95rem;
		color: var(--text-main);
	}

	.title-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.status-badge {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.15rem 0.5rem;
		border-radius: 4px;
		background: var(--bg-main);
		color: var(--text-dim);
	}

	.status-badge.active {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
	}

	.status-badge.pending {
		background: rgba(255, 255, 255, 0.05);
		color: #a1a1aa;
	}

	.status-badge.completed {
		background: rgba(16, 185, 129, 0.2);
		color: #10b981;
	}

	.status-badge.error {
		background: rgba(239, 68, 68, 0.2);
		color: #ef4444;
	}

	.item-sub {
		font-size: 0.8rem;
		color: var(--text-dim);
	}

	.item-progress {
		height: 4px;
		background: var(--bg-main);
		border-radius: 2px;
		margin-top: 0.4rem;
		overflow: hidden;
	}

	.item-progress .bar {
		height: 100%;
		background: var(--primary);
	}

	.item-actions {
		display: flex;
		gap: 0.5rem;
	}

	.action-icon-btn {
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		color: var(--text-dim);
		width: 34px;
		height: 34px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s;
	}

	.action-icon-btn:hover {
		color: var(--primary);
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}

	.input-field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.input-field label,
	.iframe-field .field-label {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.iframe-field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.input-field input {
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid var(--border-main);
		padding: 0.75rem 1rem;
		border-radius: 8px;
		color: var(--text-main);
		font-size: 0.95rem;
		outline: none;
		transition: border-color 0.2s;
	}

	.input-field input:focus {
		border-color: var(--text-main);
	}

	:global(.spin) {
		animation: spin 2s linear infinite;
	}

	@keyframes spin {
		from {
			transform: rotate(0deg);
		}
		to {
			transform: rotate(360deg);
		}
	}
</style>
