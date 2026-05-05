<script lang="ts">
	import { api, type PluginManifest } from '$lib/services/api';
	import {
		Upload,
		Trash2,
		ShieldCheck,
		ExternalLink,
		FileCode,
		ToggleLeft,
		ToggleRight
	} from 'lucide-svelte';
	import { clsx } from 'clsx';
	import Modal from '../components/Modal.svelte';
	import BasePage from '../components/base/BasePage.svelte';

	interface Props {
		plugins: PluginManifest[];
		onRefresh: () => void;
	}
	let { plugins, onRefresh }: Props = $props();

	let loading = $state(false);
	let uploadInput: HTMLInputElement;

	let showDeleteModal = $state(false);
	let pluginToDelete = $state('');

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

	async function togglePlugin(name: string) {
		loading = true;
		try {
			await api.togglePlugin(name);
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

	<div class="plugins-grid">
		{#each plugins as plugin (plugin.name)}
			<div class={clsx('plugin-card', !plugin.is_enabled && 'disabled')}>
				<div class="card-header">
					<div class="icon-box">
						<FileCode
							size={24}
							color={plugin.is_enabled ? 'var(--primary)' : 'var(--text-muted)'}
						/>
					</div>
					<div class="plugin-info">
						<h3>{plugin.name}</h3>
						<div class="status-tags">
							<span class={clsx('tag', plugin.is_enabled ? 'active' : 'inactive')}>
								{plugin.is_enabled ? 'Active' : 'Disabled'}
							</span>
							{#if plugin.permissions.length > 0}
								<span class="tag perm">
									<ShieldCheck size={12} />
									{plugin.permissions.length} Perms
								</span>
							{/if}
						</div>
					</div>
					<div class="card-actions">
						<button
							class="toggle-btn"
							onclick={() => togglePlugin(plugin.name)}
							title={plugin.is_enabled ? 'Disable Plugin' : 'Enable Plugin'}
						>
							{#if plugin.is_enabled}
								<ToggleRight size={24} color="var(--primary)" />
							{:else}
								<ToggleLeft size={24} color="var(--text-muted)" />
							{/if}
						</button>
						<button class="delete-btn" onclick={() => confirmDelete(plugin.name)}>
							<Trash2 size={18} />
						</button>
					</div>
				</div>

				<div class="card-body">
					<div class="capabilities">
						<header>Capabilities</header>
						<ul>
							{#each plugin.tabs as tab (tab.label)}
								<li>
									<ExternalLink size={12} />
									{tab.label} ({tab.section_id || 'uncategorized'})
								</li>
							{/each}
							{#if plugin.settings_groups.length > 0}
								<li>
									<ExternalLink size={12} />
									{plugin.settings_groups.length} Settings Groups
								</li>
							{/if}
						</ul>
					</div>
				</div>
			</div>
		{/each}
	</div>

	<Modal show={showDeleteModal} title="Delete Plugin" onClose={() => (showDeleteModal = false)}>
		<div class="modal-confirm">
			<p>
				Are you sure you want to permanently delete the plugin <strong>{pluginToDelete}</strong>?
				This action cannot be undone.
			</p>
			<div class="modal-actions">
				<button class="modal-btn secondary" onclick={() => (showDeleteModal = false)}>Cancel</button
				>
				<button class="modal-btn danger" onclick={handleDelete}>Delete Permanently</button>
			</div>
		</div>
	</Modal>
</BasePage>

<style>
	.plugins-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.5rem;
	}

	.plugin-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 16px;
		padding: 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
		transition: all 0.2s ease;
	}

	.plugin-card:hover {
		border-color: var(--primary);
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.plugin-card.disabled {
		opacity: 0.7;
	}

	.card-header {
		display: flex;
		gap: 1rem;
		align-items: flex-start;
	}

	.icon-box {
		background: rgba(var(--primary-rgb), 0.1);
		padding: 0.75rem;
		border-radius: 12px;
	}

	.plugin-info {
		flex: 1;
	}

	.plugin-info h3 {
		margin: 0 0 0.5rem;
		font-size: 1.1rem;
		font-weight: 700;
		text-transform: capitalize;
	}

	.status-tags {
		display: flex;
		gap: 0.5rem;
	}

	.tag {
		font-size: 0.65rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.2rem 0.5rem;
		border-radius: 4px;
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.tag.active {
		background: #10b98120;
		color: #10b981;
	}
	.tag.inactive {
		background: #6b728020;
		color: #6b7280;
	}
	.tag.perm {
		background: #3b82f620;
		color: #3b82f6;
	}

	.card-actions {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.toggle-btn,
	.delete-btn {
		background: none;
		border: none;
		color: var(--text-muted);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 8px;
		transition: all 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.delete-btn:hover {
		color: #ef4444;
		background: #ef444410;
	}

	.toggle-btn:hover {
		background: rgba(var(--primary-rgb), 0.1);
	}

	.capabilities header {
		font-size: 0.7rem;
		font-weight: 700;
		color: var(--text-muted);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.75rem;
	}

	.capabilities ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.capabilities li {
		font-size: 0.8rem;
		color: var(--text-main);
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0 1.25rem;
		height: 40px;
		border-radius: 10px;
		font-weight: 700;
		font-size: 0.875rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-primary:hover {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	.btn-primary:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.modal-confirm p {
		color: var(--text-muted);
		line-height: 1.6;
		margin-bottom: 2rem;
	}
	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
	}
	.modal-btn {
		padding: 0.6rem 1.2rem;
		border-radius: 8px;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		border: none;
		transition: all 0.2s;
	}
	.modal-btn.secondary {
		background: var(--bg-surface-hover);
		color: var(--text-main);
	}
	.modal-btn.danger {
		background: var(--accent);
		color: white;
	}
	.modal-btn:hover {
		opacity: 0.9;
		transform: translateY(-1px);
	}
</style>
