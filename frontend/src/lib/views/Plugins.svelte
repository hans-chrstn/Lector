<script lang="ts">
	import { api, type PluginManifest } from '$lib/services/api';
	import { toast } from '$lib/services/toast.svelte';
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
		GripVertical,
		List,
		Grid3X3
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
	let viewMode = $state<'grid' | 'list'>('grid');

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
			toast.success('Plugin order saved');
		} catch {
			toast.error('Failed to save plugin order');
		}
	}

	let loading = $state(false);
	let uploadInput: HTMLInputElement;
	let showDeleteModal = $state(false);
	let pluginToDelete = $state('');

	let showUploadModal = $state(false);
	let selectedFile = $state<File | null>(null);
	let uploadPluginName = $state('');

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

	async function handleFileSelect(e: Event) {
		const file = (e.target as HTMLInputElement).files?.[0];
		if (!file) return;
		selectedFile = file;

		try {
			await api.uploadPlugin(file);
			await onRefresh();
			toast.success('Plugin installed successfully');
			selectedFile = null;
		} catch (err: any) {
			if (err.message && err.message.includes('409')) {
				uploadPluginName = file.name.replace(/\.[^/.]+$/, "").toLowerCase();
				if (uploadPluginName === 'init') uploadPluginName = '';
				showUploadModal = true;
			} else {
				toast.error(err.message || 'Failed to install plugin');
			}
		}
	}


	async function handleUpload() {
		if (!selectedFile) return;
		showUploadModal = false;
		loading = true;
		try {
			await api.uploadPlugin(selectedFile, uploadPluginName);
			await onRefresh();
			toast.success('Plugin installed successfully');
			selectedFile = null;
			uploadPluginName = '';
		} catch (err: any) {
			toast.error(err.message || 'Failed to install plugin');
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
			toast.success('Plugin deleted');
		} catch {
			toast.error('Failed to delete plugin');
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
			toast.success(`Plugin ${name} updated`);
		} catch {
			toast.error('Failed to update plugin status');
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
		<div class="header-actions">
			<button
				class="icon-btn"
				onclick={() => (viewMode = viewMode === 'grid' ? 'list' : 'grid')}
				title={viewMode === 'grid' ? 'Switch to List' : 'Switch to Grid'}
			>
				{#if viewMode === 'grid'}
					<List size={20} />
				{:else}
					<Grid3X3 size={20} />
				{/if}
			</button>
			<button class="btn-primary" onclick={() => uploadInput.click()} disabled={loading}>
				<Upload size={18} />
				<span>Install Plugin</span>
			</button>
		</div>
		<input
			type="file"
			accept=".lua,.zip"
			bind:this={uploadInput}
			onchange={handleFileSelect}
			hidden
		/>
	{/snippet}

	<div
		class={clsx(viewMode === 'grid' ? 'plugins-grid' : 'plugins-list')}
		use:dndzone={{
			items,
			flipDurationMs: 300,
			dragDisabled: loading,
			type: 'plugins',
			dropTargetStyle: { outline: '2px dashed var(--primary)', borderRadius: '24px' }
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
						<div class="title-row">
							<h3>{plugin.name}</h3>
							{#if plugin.is_verified}
								<span title="Official Verified Plugin">
									<ShieldCheck size={16} class="text-verified" />
								</span>
							{/if}
						</div>
						<div class="status-tags">
							<span class={clsx('tag', plugin.is_enabled ? 'active' : 'inactive')}>
								{plugin.is_enabled ? 'Active' : 'Disabled'}
							</span>
							{#if !plugin.is_verified}
								<span class="tag unverified">Unverified</span>
							{/if}
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
						<button
							class="delete-btn"
							onclick={() => confirmDelete(plugin.name)}
							title="Delete Plugin"
						>
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

	<Modal show={showUploadModal} title="Install Plugin" onClose={() => (showUploadModal = false)}>
		<div class="modal-form">
			<p>Confirm the identity of the plugin before installation.</p>
			<div class="input-group">
				<label for="pname">Plugin Identifier</label>
				<input id="pname" type="text" bind:value={uploadPluginName} placeholder="my-plugin" />
			</div>
			<div class="file-info-badge">
				<FileCode size={14} />
				<span>{selectedFile?.name}</span>
			</div>
			<div class="modal-actions">
				<button class="modal-btn secondary-btn" onclick={() => (showUploadModal = false)}>
					Cancel
				</button>
				<button class="modal-btn primary-glow" onclick={handleUpload} disabled={!uploadPluginName}>
					Confirm and Install
				</button>
			</div>
		</div>
	</Modal>

	<Modal
		show={showPermissionModal}
		title="Plugin Permissions Review"
		onClose={() => (showPermissionModal = false)}
	>
		{#if pendingPlugin}
			<div class="permission-review" in:fade>
				<div class="alert-box">
					<ShieldAlert size={20} />
					<p>Review the requested capabilities before enabling this plugin.</p>
				</div>

				<div class="review-grid">
					<div class="review-section">
						<h4>Required Capabilities</h4>
						<div class="cap-list">
							{#each pendingPlugin.capabilities || [] as cap (cap)}
								{@const Icon = capabilityIcons[cap] || ShieldCheck}
								<div class="cap-item">
									<Icon size={16} />
									<span>{cap}</span>
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

				<div class="disclaimer">
					By enabling this plugin, you acknowledge that it will have access to the resources listed
					above. Only enable plugins from sources you trust.
				</div>

				<div class="modal-actions">
					<button class="modal-btn secondary-btn" onclick={() => (showPermissionModal = false)}>
						Cancel
					</button>
					<button class="modal-btn primary-glow" onclick={() => executeToggle(pendingPlugin!.name)}>
						Enable Plugin
					</button>
				</div>
			</div>
		{/if}
	</Modal>

	<Modal show={showDeleteModal} title="Delete Plugin" onClose={() => (showDeleteModal = false)}>
		<div class="delete-confirm">
			<p>Are you sure you want to delete <strong>{pluginToDelete}</strong>?</p>
			<div class="modal-actions">
				<button class="modal-btn secondary-btn" onclick={() => (showDeleteModal = false)}>
					Cancel
				</button>
				<button class="modal-btn danger-glow" onclick={handleDelete}>Delete Permanently</button>
			</div>
		</div>
	</Modal>
</BasePage>

<style>
	.header-actions {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.icon-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
		width: 40px;
		height: 40px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s;
	}

	.icon-btn:hover {
		border-color: var(--border-bright);
		color: var(--text-main);
		background: var(--bg-surface-hover);
	}

	.plugins-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 1.5rem;
		min-height: 100px;
	}

	.plugins-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.plugin-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 24px;
		overflow: hidden;
		transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
		display: flex;
		flex-direction: column;
	}

	.plugins-list .plugin-card {
		flex-direction: row;
		flex-wrap: wrap;
	}

	.plugin-card.is-disabled {
		opacity: 0.6;
		filter: grayscale(0.4);
	}

	.card-header {
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1.25rem;
		background: rgba(255, 255, 255, 0.01);
		border-bottom: 1px solid var(--border-main);
	}

	.plugins-list .card-header {
		border-bottom: none;
		flex: 1;
		min-width: 300px;
	}

	.grab-handle {
		color: var(--text-dim);
		cursor: grab;
		padding: 0.5rem;
		margin-left: -0.5rem;
		opacity: 0.5;
		transition: opacity 0.2s;
	}

	.grab-handle:hover {
		opacity: 1;
	}

	.icon-box {
		width: 48px;
		height: 48px;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.plugin-info {
		flex: 1;
		min-width: 0;
	}

	.title-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.25rem;
	}

	.plugin-info h3 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 700;
		text-transform: capitalize;
		color: var(--text-main);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.text-verified {
		color: #3b82f6;
	}

	.status-tags {
		display: flex;
		gap: 0.5rem;
	}

	.tag {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.15rem 0.5rem;
		border-radius: 6px;
	}

	.tag.active {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
		border: 1px solid rgba(16, 185, 129, 0.2);
	}

	.tag.inactive {
		background: var(--bg-main);
		color: var(--text-dim);
		border: 1px solid var(--border-main);
	}

	.tag.unverified {
		background: #f59e0b15;
		color: #f59e0b;
		border: 1px solid #f59e0b30;
	}

	.card-actions {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.switch-track {
		width: 44px;
		height: 24px;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		position: relative;
		cursor: pointer;
		transition: all 0.2s;
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
		transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
	}

	.switch-track.active .switch-thumb {
		transform: translateX(20px);
	}

	.delete-btn {
		background: none;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 10px;
		transition: all 0.2s;
	}

	.delete-btn:hover {
		color: #ef4444;
		background: rgba(239, 68, 68, 0.1);
	}

	.card-body {
		padding: 1.25rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		flex: 1;
	}

	.plugins-list .card-body {
		border-left: 1px solid var(--border-main);
		padding: 1rem 1.5rem;
	}

	.capabilities-section header,
	.permissions-section header {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-dim);
		margin-bottom: 0.75rem;
	}

	.cap-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.cap-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4rem 0.75rem;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		border-radius: 10px;
		color: var(--text-muted);
		font-size: 0.75rem;
		font-weight: 600;
	}

	.perm-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
	}

	.perm-chip {
		font-size: 0.75rem;
		font-family: var(--font-mono);
		color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
		padding: 0.2rem 0.6rem;
		border-radius: 6px;
	}

	.modal-form {
		padding: 1rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.modal-form p {
		margin: 0;
		font-size: 0.9375rem;
		color: var(--text-dim);
	}

	.input-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.input-group label {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		color: var(--text-dim);
	}

	.input-group input {
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.75rem 1rem;
		border-radius: 12px;
		font-size: 1rem;
		outline: none;
	}

	.input-group input:focus {
		border-color: var(--primary);
	}

	.file-info-badge {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		background: rgba(var(--primary-rgb), 0.05);
		border: 1px solid rgba(var(--primary-rgb), 0.1);
		border-radius: 12px;
		color: var(--primary);
		font-size: 0.875rem;
		font-weight: 600;
	}

	.modal-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		margin-top: 1rem;
	}

	.delete-confirm {
		padding: 1rem;
		text-align: center;
	}

	.delete-confirm p {
		margin-bottom: 2rem;
		color: var(--text-main);
		font-size: 1rem;
		line-height: 1.6;
	}

	.modal-btn {
		padding: 0.75rem 1.5rem;
		border-radius: 14px;
		font-weight: 700;
		font-size: 0.875rem;
		cursor: pointer;
		border: none;
		transition: all 0.2s;
	}

	.primary-glow {
		background: white;
		color: black;
		box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
	}

	.primary-glow:hover:not(:disabled) {
		background: #f4f4f5;
		transform: translateY(-1px);
		box-shadow: 0 0 20px rgba(255, 255, 255, 0.2);
	}

	.primary-glow:disabled {
		opacity: 0.5;
		cursor: not-allowed;
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

	@media (max-width: 1000px) {
		.plugins-grid {
			grid-template-columns: 1fr;
		}
		.plugins-list .plugin-card {
			flex-direction: column;
		}
		.plugins-list .card-body {
			border-left: none;
			border-top: 1px solid var(--border-main);
		}
	}
</style>
