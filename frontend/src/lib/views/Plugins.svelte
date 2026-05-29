<script lang="ts">
	import { api, type PluginManifest } from '$lib/services/api';
	import {
		Upload,
		Trash2,
		ShieldCheck,
		FileCode,
		ShieldAlert,
		CheckCircle2,
		Globe,
		Database,
		Layout,
		History,
		Zap,
		GripVertical
	} from 'lucide-svelte';
	import { clsx } from 'clsx';
	import { fade } from 'svelte/transition';
	import { untrack } from 'svelte';
	import { dndzone } from 'svelte-dnd-action';
	import Modal from '../components/Modal.svelte';
	import BasePage from '../components/base/BasePage.svelte';
	interface Props {
		plugins: PluginManifest[];
		onRefresh: () => void;
	}
	let { plugins, onRefresh }: Props = $props();
	let items = $state<any[]>([]);
	$effect(() => {
		if (plugins && plugins.length !== untrack(() => items.length)) {
			items = plugins.map((p) => ({ ...p, id: p.name }));
		}
	});
	function handleDndConsider(e: CustomEvent<any>) {
		items = e.detail.items;
	}
	async function handleDndFinalize(e: CustomEvent<any>) {
		const newItems = e.detail.items;
		items = newItems;
		try {
			await api.reorderPlugins(newItems.map((i: any) => i.name));
			await onRefresh();
		} catch (err) {
			console.error('Failed to persist plugin order:', err);
		}
	}
	let loading = $state(false);
	let uploadInput: HTMLInputElement;
	let showDeleteModal = $state(false);
	let pluginToDelete = $state('');
	let showPermissionModal = $state(false);
	let pendingPlugin = $state<PluginManifest | null>(null);
	const capabilityIcons: Record<string, any> = {
		network: Globe,
		storage: Database,
		ui: Layout,
		source: Zap,
		background: History,
		interaction: CheckCircle2
	};
	async function handleUpload(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		loading = true;
		try {
			await api.uploadPlugin(file);
			await onRefresh();
		} finally {
			loading = false;
		}
	}
	async function confirmDelete(name: string) {
		pluginToDelete = name;
		showDeleteModal = true;
	}
	async function handleDelete() {
		showDeleteModal = false;
		loading = true;
		try {
			await api.deletePlugin(pluginToDelete);
			await onRefresh();
		} finally {
			loading = false;
		}
	}
	async function togglePlugin(plugin: PluginManifest) {
		if (!plugin.is_enabled) {
			pendingPlugin = plugin;
			showPermissionModal = true;
			return;
		}
		await executeToggle(plugin.name);
	}
	async function executeToggle(name: string) {
		showPermissionModal = false;
		loading = true;
		items = items.map((item) =>
			item.name === name ? { ...item, is_enabled: !item.is_enabled } : item
		);
		try {
			await api.togglePlugin(name);
			await onRefresh();
		} catch (err) {
			console.error('Failed to toggle plugin:', err);
			await onRefresh();
		} finally {
			loading = false;
		}
	}
</script>

<BasePage
	title="Plugins"
	titleBadge="{plugins.length} Installed"
	subtitle="Extend your instance with new sources and capabilities"
>
	{#snippet actions()}
		<button class="btn-primary" onclick={() => uploadInput.click()} disabled={loading}>
			<Upload size={18} />
			<span>Install Plugin</span>
		</button>
		<input type="file" accept=".lua" bind:this={uploadInput} onchange={handleUpload} hidden />
	{/snippet}
	<div
		class="plugins-grid"
		use:dndzone={{
			items,
			flipDurationMs: 300,
			dragDisabled: loading,
			type: 'plugins',
			dropTargetStyle: { outline: '2px dashed var(--primary)', borderRadius: '20px' }
		}}
		onconsider={handleDndConsider}
		onfinalize={handleDndFinalize}
	>
		{#each items as plugin (plugin.id)}
			<div class={clsx('plugin-card', !plugin.is_enabled && 'is-disabled')}>
				<div class="card-header">
					<div class="grab-handle">
						<GripVertical size={20} />
					</div>
					<div class="icon-box">
						<FileCode size={24} class={plugin.is_enabled ? 'text-primary' : 'text-muted'} />
					</div>
					<div class="plugin-info">
						<h3>{plugin.name}</h3>
						<div class="status-tags">
							<span class={clsx('tag', plugin.is_enabled ? 'active' : 'inactive')}>
								{plugin.is_enabled ? 'Active' : 'Disabled'}
							</span>
						</div>
					</div>
					<div class="card-actions">
						<div
							class={clsx('switch-track', plugin.is_enabled && 'active')}
							role="button"
							tabindex="0"
							onclick={() => togglePlugin(plugin)}
							onkeydown={(e) => e.key === 'Enter' && togglePlugin(plugin)}
						>
							<div class="switch-thumb"></div>
						</div>
						<button class="delete-btn" onclick={() => confirmDelete(plugin.name)}>
							<Trash2 size={18} />
						</button>
					</div>
				</div>
				<div class="card-body">
					<div class="capabilities-section">
						<header>Requested Capabilities</header>
						<div class="cap-list">
							{#each plugin.capabilities || [] as cap (cap)}
								{@const Icon = capabilityIcons[cap] || ShieldCheck}
								<div class="cap-item" title="Requires {cap} access">
									<Icon size={14} />
									<span>{cap}</span>
								</div>
							{/each}
						</div>
					</div>
					{#if plugin.permissions && plugin.permissions.length > 0}
						<div class="permissions-section">
							<header>Domains</header>
							<div class="perm-list">
								{#each plugin.permissions as perm (perm)}
									<span class="perm-chip">{perm}</span>
								{/each}
							</div>
						</div>
					{/if}
				</div>
			</div>
		{/each}
	</div>
	<Modal
		show={showPermissionModal}
		title="Plugin Permissions Review"
		onClose={() => (showPermissionModal = false)}
	>
		{#if pendingPlugin}
			<div class="permission-review" in:fade>
				<div class="alert-box">
					<ShieldAlert size={20} />
					<p>Enable <strong>{pendingPlugin.name}</strong>? Review the requested access below:</p>
				</div>
				<div class="review-grid">
					<div class="review-section">
						<h4>Capabilities</h4>
						<div class="cap-list">
							{#each pendingPlugin.capabilities || [] as cap (cap)}
								{@const Icon = capabilityIcons[cap] || ShieldCheck}
								<div class="cap-item">
									<Icon size={16} />
									<span>{cap} access</span>
								</div>
							{/each}
						</div>
					</div>
					{#if pendingPlugin.permissions && pendingPlugin.permissions.length > 0}
						<div class="review-section">
							<h4>Network Domains</h4>
							<div class="perm-list">
								{#each pendingPlugin.permissions as perm (perm)}
									<span class="perm-chip">{perm}</span>
								{/each}
							</div>
						</div>
					{/if}
				</div>
				<p class="disclaimer">
					By enabling this plugin, you grant it permission to use the capabilities and access the
					domains listed above. Plugins run in a secure sandbox but may still perform the actions
					they have requested.
				</p>
				<div class="modal-actions">
					<button class="modal-btn secondary" onclick={() => (showPermissionModal = false)}>
						Cancel
					</button>
					<button
						class="modal-btn primary-glow"
						onclick={() => pendingPlugin && executeToggle(pendingPlugin.name)}
					>
						Understand and Enable
					</button>
				</div>
			</div>
		{/if}
	</Modal>
	<Modal show={showDeleteModal} title="Delete Plugin" onClose={() => (showDeleteModal = false)}>
		<div class="modal-confirm">
			<p>
				Are you sure you want to permanently delete the plugin <strong>{pluginToDelete}</strong>?
				This action cannot be undone.
			</p>
			<div class="modal-actions">
				<button class="modal-btn secondary" onclick={() => (showDeleteModal = false)}>Cancel</button
				>
				<button class="modal-btn danger-glow" onclick={handleDelete}>Delete Permanently</button>
			</div>
		</div>
	</Modal>
</BasePage>

<style>
	.plugins-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
		gap: 1.5rem;
	}
	.plugin-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 20px;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		position: relative;
		overflow: hidden;
	}
	.plugin-card:hover {
		border-color: var(--primary);
		transform: translateY(-4px);
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
	}
	.grab-handle {
		cursor: grab;
		color: var(--text-dim);
		padding: 0.5rem;
		margin-left: -0.75rem;
		transition: color 0.2s;
	}
	.grab-handle:hover {
		color: var(--text-main);
	}
	.grab-handle:active {
		cursor: grabbing;
	}
	.plugin-card.is-disabled {
		opacity: 0.5;
		filter: grayscale(0.5);
	}
	.card-header {
		display: flex;
		gap: 1rem;
		align-items: center;
	}
	.icon-box {
		background: rgba(255, 255, 255, 0.03);
		padding: 0.75rem;
		border-radius: 14px;
		border: 1px solid var(--border-main);
	}
	.plugin-info {
		flex: 1;
	}
	.plugin-info h3 {
		margin: 0 0 0.25rem;
		font-size: 1.1rem;
		font-weight: 700;
		text-transform: capitalize;
		color: var(--text-main);
	}
	.status-tags {
		display: flex;
		gap: 0.5rem;
	}
	.tag {
		font-size: 0.6rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.2rem 0.6rem;
		border-radius: 6px;
	}
	.tag.active {
		background: #10b98120;
		color: #10b981;
	}
	.tag.inactive {
		background: var(--bg-surface-hover);
		color: var(--text-muted);
	}
	.card-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	.switch-track {
		width: 44px;
		height: 24px;
		background: var(--bg-surface-hover);
		border-radius: 100px;
		position: relative;
		cursor: pointer;
		transition: all 0.3s;
		border: 1px solid var(--border-main);
	}
	.switch-track.active {
		background: var(--primary);
		border-color: var(--primary);
	}
	.switch-thumb {
		width: 18px;
		height: 18px;
		background: white;
		border-radius: 50%;
		position: absolute;
		top: 2px;
		left: 2px;
		transition: all 0.3s cubic-bezier(0.175, 0.885, 0.32, 1.275);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}
	.switch-track.active .switch-thumb {
		left: 22px;
	}
	.delete-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 10px;
		transition: all 0.2s;
	}
	.delete-btn:hover {
		color: #ef4444;
		background: #ef444415;
	}
	.capabilities-section {
		margin-top: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.permissions-section {
		margin-top: 0.75rem;
	}

	.capabilities-section header,
	.permissions-section header {
		font-size: 0.65rem;
		font-weight: 800;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin-bottom: 0.3rem;
	}
	.cap-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	.cap-item {
		background: rgba(255, 255, 255, 0.04);
		border: 1px solid var(--border-main);
		padding: 0.4rem 0.7rem;
		border-radius: 10px;
		font-size: 0.75rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--text-main);
	}
	.perm-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
	}
	.perm-chip {
		font-size: 0.7rem;
		font-family: 'Geist Mono', monospace;
		color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
		padding: 0.2rem 0.5rem;
		border-radius: 6px;
	}
	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 1rem;
	}
	.modal-btn {
		padding: 0.75rem 1.5rem;
		border-radius: 12px;
		font-size: 0.875rem;
		font-weight: 700;
		cursor: pointer;
		border: none;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	.modal-btn.secondary {
		background: var(--bg-surface-hover);
		color: var(--text-main);
		border: 1px solid var(--border-main);
	}
	.modal-btn.secondary:hover {
		background: var(--border-main);
		border-color: var(--text-dim);
	}
	.primary-glow {
		background: white;
		color: black;
		box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
	}
	.primary-glow:hover {
		background: #f4f4f5;
		transform: translateY(-1px);
		box-shadow: 0 0 20px rgba(255, 255, 255, 0.2);
	}
	.modal-btn.danger-glow {
		background: #ef4444;
		color: white;
		box-shadow: 0 4px 12px rgba(239, 68, 68, 0.2);
	}
	.modal-btn.danger-glow:hover {
		background: #dc2626;
		transform: translateY(-1px);
		box-shadow: 0 0 20px rgba(239, 68, 68, 0.3);
	}
	.permission-review {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}
	.alert-box {
		background: #f59e0b15;
		color: #f59e0b;
		padding: 1rem;
		border-radius: 12px;
		display: flex;
		gap: 0.75rem;
		align-items: center;
		border: 1px solid #f59e0b30;
	}
	.alert-box p {
		margin: 0;
		font-size: 0.9rem;
		font-weight: 500;
	}
	.review-grid {
		display: grid;
		gap: 1.5rem;
	}
	.review-section h4 {
		margin: 0 0 0.75rem;
		font-size: 0.8rem;
		font-weight: 700;
		color: var(--text-muted);
		text-transform: uppercase;
	}
	.disclaimer {
		font-size: 0.75rem;
		color: var(--text-dim);
		line-height: 1.6;
		padding: 1rem;
		background: var(--bg-surface-hover);
		border-radius: 12px;
	}
	.btn-primary {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0 1.25rem;
		height: 40px;
		border-radius: 12px;
		font-weight: 700;
		font-size: 0.875rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		transition: all 0.2s;
	}
	.btn-primary:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
	}
	.text-primary {
		color: var(--primary);
	}
	.text-muted {
		color: var(--text-muted);
	}
</style>
