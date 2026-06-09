<script lang="ts">
	import ShieldAlert from 'lucide-svelte/icons/shield-alert';
	import Library from 'lucide-svelte/icons/library';
	import FolderPlus from 'lucide-svelte/icons/folder-plus';
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import Loader2 from 'lucide-svelte/icons/loader-2';
	import HardDrive from 'lucide-svelte/icons/hard-drive';
	import LayoutGrid from 'lucide-svelte/icons/layout-grid';
	import { onMount } from 'svelte';
	import { api } from '$lib/services/api';
	import BasePage from '../components/base/BasePage.svelte';
	import DynamicSettings from '../components/DynamicSettings.svelte';
	import type { PluginManifest } from '$lib/services/api';

	interface Props {
		pluginManifests?: PluginManifest[];
	}
	let { pluginManifests = [] }: Props = $props();

	let libraryPaths = $state<{ id: number; path: string; pattern: string; is_system: boolean }[]>(
		[]
	);
	let newPath = $state('');
	let newPattern = $state('None/Flat');
	let customPattern = $state('');
	let loading = $state(false);
	let scanning = $state(false);

	const patterns = [
		'None/Flat',
		'/{Author}/{Title}',
		'/{Author}/{Group}/{Title}',
		'/{Group}/{Title}',
		'Custom...'
	];

	async function fetchPaths() {
		try {
			libraryPaths = await api.getLibraryPaths();
		} catch {
			console.error('Failed to fetch library paths');
		}
	}

	async function handleAddPath() {
		if (!newPath) return;
		loading = true;
		try {
			const finalPattern = newPattern === 'Custom...' ? customPattern : newPattern;
			await api.addLibraryPath(newPath, finalPattern);
			newPath = '';
			customPattern = '';
			await fetchPaths();
		} finally {
			loading = false;
		}
	}

	async function handleDeletePath(id: number) {
		try {
			await api.deleteLibraryPath(id);
			await fetchPaths();
		} catch {
			console.error('Failed to fetch library paths');
		}
	}

	async function handleScan() {
		scanning = true;
		try {
			await api.scanLibrary();
			setTimeout(() => {
				scanning = false;
			}, 2000);
		} catch {
			console.error('Scan failed');
			scanning = false;
		}
	}

	onMount(fetchPaths);
</script>

<BasePage title="Settings" subtitle="Manage your library sources and instance configuration">
	<div class="settings-container">
		<section class="settings-section">
			<div class="section-header">
				<div class="header-info">
					<Library size={24} class="text-primary" />
					<div class="title-stack">
						<h2>Library Folders</h2>
						<p>Directories Lector monitors for automatic book imports</p>
					</div>
				</div>
				<button class="scan-action-btn" onclick={handleScan} disabled={scanning}>
					{#if scanning}
						<Loader2 size={18} class="spin" />
					{:else}
						<RefreshCw size={18} />
					{/if}
					<span>Scan Libraries</span>
				</button>
			</div>

			<div class="add-library-panel">
				<div class="panel-inputs">
					<div class="input-field">
						<label for="path">Server Path</label>
						<div class="input-wrapper">
							<HardDrive size={16} class="input-icon" />
							<input id="path" type="text" placeholder="/home/user/library" bind:value={newPath} />
						</div>
					</div>
					<div class="input-field">
						<label for="pattern">Metadata Pattern</label>
						<div class="input-wrapper">
							<LayoutGrid size={16} class="input-icon" />
							<select id="pattern" bind:value={newPattern}>
								{#each patterns as p (p)}
									<option value={p}>{p}</option>
								{/each}
							</select>
						</div>
					</div>
					{#if newPattern === 'Custom...'}
						<div class="input-field">
							<label for="custom-pattern">Define Pattern</label>
							<div class="input-wrapper">
								<LayoutGrid size={16} class="input-icon" />
								<input
									id="custom-pattern"
									type="text"
									placeholder="/&#123;Author&#125;/&#123;Title&#125;"
									bind:value={customPattern}
								/>
							</div>
						</div>
					{/if}
				</div>
				<button class="add-library-btn" onclick={handleAddPath} disabled={loading || !newPath}>
					<FolderPlus size={18} />
					<span>Add Location</span>
				</button>
			</div>

			<div class="library-grid">
				{#each libraryPaths as lp (lp.id)}
					<div class="library-card" class:system-managed={lp.is_system}>
						<div class="card-content">
							<div class="card-icon">
								<HardDrive size={20} />
							</div>
							<div class="card-info">
								<span class="path-label">
									{lp.is_system ? 'Internal Managed Storage' : lp.path}
								</span>
								<div class="meta-row">
									<span class="pattern-pill">{lp.pattern}</span>
									<span class="status-dot" class:system={lp.is_system}></span>
									<span class="status-text" class:system={lp.is_system}>
										{lp.is_system ? 'System Managed' : 'Monitoring'}
									</span>
								</div>
							</div>
						</div>
						{#if !lp.is_system}
							<button
								class="remove-btn"
								onclick={() => handleDeletePath(lp.id)}
								title="Remove folder"
							>
								<Trash2 size={18} />
							</button>
						{/if}
					</div>
				{/each}
			</div>
		</section>

		<section class="settings-section">
			<div class="section-header">
				<div class="header-info">
					<ShieldAlert size={24} class="text-primary" />
					<div class="title-stack">
						<h2>System Controls</h2>
						<p>Core instance maintenance and security options</p>
					</div>
				</div>
			</div>

			<div class="system-controls">
				<div class="control-item">
					<div class="control-info">
						<h3>Clear Metadata Cache</h3>
						<p>Deletes cached covers and discovery results. Forces a full re-sync.</p>
					</div>
					<button class="danger-outline-btn">Flush Cache</button>
				</div>
				<div class="control-item">
					<div class="control-info">
						<h3>Instance Identity</h3>
						<p>Production Build v0.3.5 • Running on Linux Environment</p>
					</div>
				</div>
			</div>
		</section>

		{#each pluginManifests as plugin (plugin.name)}
			{#if plugin.settings_groups && plugin.settings_groups.length > 0}
				<section class="settings-section">
					<div class="section-header">
						<div class="header-info">
							<div class="title-stack">
								<h2>{plugin.name.charAt(0).toUpperCase() + plugin.name.slice(1)} Settings</h2>
								<p>Configure options for this plugin</p>
							</div>
						</div>
					</div>
					{#each plugin.settings_groups as group (group.id)}
						<DynamicSettings pluginName={plugin.name} groupId={group.id} />
					{/each}
				</section>
			{/if}
		{/each}
	</div>
</BasePage>

<style>
	.settings-container {
		display: flex;
		flex-direction: column;
		gap: 3rem;
		max-width: 1000px;
		margin: 0 auto;
	}

	.settings-section {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		padding-bottom: 1rem;
		border-bottom: 1px solid var(--border-main);
	}

	.header-info {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.title-stack h2 {
		font-size: 1.25rem;
		font-weight: 800;
		color: var(--text-main);
		margin: 0;
		letter-spacing: -0.02em;
	}

	.title-stack p {
		font-size: 0.875rem;
		color: var(--text-dim);
		margin: 0.25rem 0 0 0;
	}

	.scan-action-btn {
		background: var(--bg-surface-hover);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.6rem 1.25rem;
		border-radius: 10px;
		font-size: 0.8125rem;
		font-weight: 700;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.scan-action-btn:hover:not(:disabled) {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.05);
	}

	.add-library-panel {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 16px;
		padding: 1.5rem;
		display: flex;
		gap: 1.5rem;
		align-items: flex-end;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
	}

	.panel-inputs {
		flex: 1;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.25rem;
	}

	.input-field {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.input-field label {
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--text-dim);
	}

	.input-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}

	.input-wrapper :global(.input-icon) {
		position: absolute;
		left: 1.25rem;
		color: var(--text-dim);
		pointer-events: none;
	}

	input[type='text'],
	select {
		width: 100%;
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.75rem 1.25rem 0.75rem 3.5rem;
		border-radius: 10px;
		font-size: 0.9375rem;
		outline: none;
		transition: all 0.2s;
		appearance: none;
	}

	select {
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%2371717a' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 1.25rem center;
		padding-right: 2.75rem;
	}

	input[type='text']:focus,
	select:focus {
		border-color: var(--primary);
		background: var(--bg-surface);
		box-shadow: 0 0 0 4px rgba(var(--primary-rgb), 0.1);
	}

	.add-library-btn {
		background: white;
		color: black;
		border: none;
		padding: 0 1.5rem;
		height: 48px;
		border-radius: 12px;
		font-weight: 800;
		font-size: 0.875rem;
		display: flex;
		align-items: center;
		gap: 0.6rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.add-library-btn:hover:not(:disabled) {
		background: #f4f4f5;
		transform: scale(1.02);
	}

	.library-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 1rem;
	}

	.library-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 14px;
		padding: 1.25rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: all 0.2s;
	}

	.library-card:hover {
		border-color: var(--primary);
		background: var(--bg-surface-hover);
	}

	.library-card.system-managed {
		border-style: dashed;
		opacity: 0.8;
	}

	.library-card.system-managed:hover {
		opacity: 1;
	}

	.card-content {
		display: flex;
		gap: 1rem;
		align-items: center;
		overflow: hidden;
	}

	.card-icon {
		width: 40px;
		height: 40px;
		background: rgba(var(--primary-rgb), 0.1);
		color: var(--primary);
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.card-info {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		overflow: hidden;
	}

	.path-label {
		font-size: 0.875rem;
		font-family: 'Geist Mono', monospace;
		font-weight: 600;
		color: var(--text-main);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.meta-row {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.pattern-pill {
		font-size: 0.7rem;
		font-weight: 800;
		background: var(--bg-main);
		color: var(--text-dim);
		padding: 0.15rem 0.6rem;
		border-radius: 5px;
		border: 1px solid var(--border-main);
	}

	.status-dot {
		width: 6px;
		height: 6px;
		background: #10b981;
		border-radius: 50%;
		box-shadow: 0 0 8px rgba(16, 185, 129, 0.4);
	}

	.status-dot.system {
		background: #3b82f6;
		box-shadow: 0 0 8px rgba(59, 130, 246, 0.4);
	}

	.status-text {
		font-size: 0.75rem;
		font-weight: 600;
		color: #10b981;
	}

	.status-text.system {
		color: #3b82f6;
	}

	.remove-btn {
		background: transparent;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		padding: 0.6rem;
		border-radius: 10px;
		transition: all 0.2s;
	}

	.remove-btn:hover {
		color: #ef4444;
		background: rgba(239, 68, 68, 0.1);
	}

	.system-controls {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 16px;
		padding: 1rem;
		display: flex;
		flex-direction: column;
	}

	.control-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.25rem;
	}

	.control-item:not(:last-child) {
		border-bottom: 1px solid var(--border-main);
	}

	.control-info h3 {
		font-size: 0.9375rem;
		font-weight: 700;
		color: var(--text-main);
		margin: 0;
	}

	.control-info p {
		font-size: 0.8125rem;
		color: var(--text-dim);
		margin: 0.25rem 0 0 0;
	}

	.danger-outline-btn {
		background: transparent;
		border: 1px solid #ef4444;
		color: #ef4444;
		padding: 0.5rem 1rem;
		border-radius: 8px;
		font-size: 0.8125rem;
		font-weight: 700;
		cursor: pointer;
		transition: all 0.2s;
	}

	.danger-outline-btn:hover {
		background: #ef4444;
		color: white;
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
		.add-library-panel {
			flex-direction: column;
			align-items: stretch;
		}
		.panel-inputs {
			grid-template-columns: 1fr;
		}
		.library-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
