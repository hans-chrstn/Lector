<script lang="ts">
	import { onMount } from 'svelte';
	import Menu from 'lucide-svelte/icons/menu';
	import X from 'lucide-svelte/icons/x';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Library from '$lib/views/Library.svelte';
	import Search from '$lib/views/Search.svelte';
	import PluginsView from '$lib/views/Plugins.svelte';
	import Explorer from '$lib/views/Explorer.svelte';
	import Detail from '$lib/views/Detail.svelte';
	import DynamicView from '$lib/views/DynamicView.svelte';
	import PluginExplorer from '$lib/views/PluginExplorer.svelte';
	import Reader from '$lib/views/Reader.svelte';
	import History from '$lib/views/History.svelte';
	import Settings from '$lib/views/Settings.svelte';
	import Modal from '$lib/components/Modal.svelte';
	import {
		api,
		type PluginManifest,
		type Group,
		type Document as LectorDocument,
		type SearchItem,
		type Chapter
	} from '$lib/services/api';
	import { baseManifest } from '$lib/services/ui';
	import { clsx } from 'clsx';
	import { toast } from '$lib/services/toast.svelte';

	let plugins = $state<PluginManifest[]>([]);
	let sources = $state<string[]>([]);
	let groups = $state<Group[]>([]);
	let documents = $state<LectorDocument[]>([]);
	let history = $state<LectorDocument[]>([]);
	let view = $state('library');
	let originNav = $state({ view: 'library', plugin: 'system', tabId: 'sys:library' });
	let currentPlugin = $state('system');
	let currentTabId = $state('sys:library');
	let loading = $state(false);
	let sidebarOpen = $state(false);

	let showCreateGroupModal = $state(false);
	let newGroupName = $state('');
	let showConfirmModal = $state(false);
	let confirmTitle = $state('');
	let confirmMessage = $state('');
	let onConfirm: (() => void) | null = $state(null);

	let searchItems = $state<SearchItem[]>([]);
	let searchQuery = $state('');
	let searchSource = $state('library');
	let popularResults = $state<SearchItem[]>([]);
	let latestResults = $state<SearchItem[]>([]);
	let activeDocument = $state<LectorDocument | null>(null);
	let activeChapter = $state<Chapter | null>(null);

	$effect(() => {
		if (typeof document !== 'undefined') {
			const oldStyles = document.querySelectorAll('style[data-plugin-style]');
			oldStyles.forEach((s) => s.remove());

			plugins.forEach((p) => {
				if (p.is_enabled && p.css) {
					const style = document.createElement('style');
					style.setAttribute('data-plugin-style', p.name);
					style.textContent = p.css;
					document.head.appendChild(style);
				}
			});
		}
	});

	onMount(async () => {
		try {
			await Promise.all([refreshPlugins(), refreshGroups(), refreshDocuments()]);
		} catch {}
	});

	async function refreshPlugins() {
		const [manifest, activeSources] = await Promise.all([
			api.getPluginsManifest(),
			api.getActivePlugins()
		]);
		plugins = [baseManifest, ...manifest];
		sources = activeSources || [];
	}

	$effect(() => {
		if (view === 'library') {
			refreshDocuments();
			refreshGroups();
		}
	});

	$effect(() => {
		if (view === 'plugins') {
			refreshPlugins();
		}
	});
	async function refreshGroups() {
		const res = await api.getGroups();
		groups = res || [];
	}
	async function refreshDocuments(archived: boolean = false) {
		const res = await api.getDocuments(archived);
		documents = res || [];
	}
	async function refreshHistory() {
		const res = await api.getHistory();
		history = res || [];
	}

	async function handleCreateGroup() {
		showCreateGroupModal = true;
		newGroupName = '';
	}

	async function submitCreateGroup() {
		if (newGroupName.trim()) {
			await api.createGroup(newGroupName);
			await refreshGroups();
			showCreateGroupModal = false;
		}
	}

	async function handleBatchDelete(ids: number[]) {
		confirmTitle = 'Delete Documents';
		confirmMessage = `Are you sure you want to permanently delete ${ids.length} documents?`;
		onConfirm = async () => {
			await api.batchDeleteDocuments(ids);
			await refreshDocuments();
			showConfirmModal = false;
		};
		showConfirmModal = true;
	}

	async function handleBatchMove(ids: number[], groupId: number) {
		await api.batchMoveDocuments(ids, groupId);
		await refreshDocuments();
	}

	async function handleBatchArchive(ids: number[], archive: boolean) {
		await api.batchArchiveDocuments(ids, archive);
		await refreshDocuments(!archive);
	}

	async function handleBatchMarkRead(ids: number[], isRead: boolean) {
		await api.batchMarkReadDocuments(ids, isRead);
		await refreshDocuments();
	}

	async function handleBatchRefresh(ids: number[]) {
		try {
			toast.success(`Refreshing ${ids.length} documents...`);
			await api.batchRefreshDocuments(ids);
			await refreshDocuments();
			toast.success('Documents refreshed successfully');
		} catch {
			toast.error('Failed to refresh documents');
		}
	}

	async function handleNavigate(targetView: string, plugin?: string, tabId?: string) {
		sidebarOpen = false;

		if (targetView === 'search' && originNav.view !== 'search') {
			searchQuery = '';
			searchSource = 'library';
			searchItems = [];
		} else if (targetView === 'search' && view === 'search') {
			searchQuery = '';
			searchSource = 'library';
			searchItems = [];
		}

		if (targetView !== 'detail' && targetView !== 'reader') {
			originNav = {
				view: targetView,
				plugin: plugin || currentPlugin,
				tabId: tabId || currentTabId
			};
		}

		currentPlugin = plugin || '';
		currentTabId = tabId || '';
		view = targetView;

		if (targetView === 'library') {
			currentPlugin = 'system';
			currentTabId = 'sys:library';
			await refreshDocuments();
		} else if (targetView === 'history') {
			loading = true;
			try {
				await refreshHistory();
			} finally {
				loading = false;
			}
		}

		if (plugin && plugin !== 'system' && targetView === 'plugin') {
			loading = true;
			try {
				const [pop, lat] = await Promise.all([
					api.getDocumentPopular(plugin),
					api.getDocumentLatest(plugin)
				]);
				popularResults = pop || [];
				latestResults = lat || [];
			} finally {
				loading = false;
			}
		}
	}
	async function handleSearch(q: string, source: string) {
		loading = true;
		try {
			if (source === 'library') {
				const res = await api.searchLibrary(q);
				searchItems = (res || []).map((doc) => ({
					title: doc.title,
					url: doc.url,
					cover_url: doc.cover_url,
					info: doc.author || doc.source,
					source: doc.source
				}));
			} else {
				const res = await api.search(source, q);
				searchItems = res.results || [];
				if (res.errors && res.errors.length > 0) {
					res.errors.forEach((err: string) => {
						toast.error(err);
					});
				}
			}
		} finally {
			loading = false;
		}
	}

	async function handleSelectDocument(url: string, source: string) {
		loading = true;
		try {
			activeDocument = await api.ensureDocument(url, source, true);
			if (view !== 'detail' && view !== 'reader') {
				originNav = { view, plugin: currentPlugin, tabId: currentTabId };
			}
			view = 'detail';
		} catch (err: any) {
			toast.error(err.message || 'Failed to load document');
		} finally {
			loading = false;
		}
	}

	async function handleToggleLibrary() {
		if (!activeDocument) return;
		const target = !activeDocument.is_in_library;
		await api.toggleLibrary(activeDocument.id, target);
		activeDocument.is_in_library = target;
		await refreshDocuments();
	}

	async function handleReadChapter(chapter: any) {
		if (!chapter) return;
		loading = true;
		try {
			activeChapter = await api.getChapter(chapter.id);
			view = 'reader';
		} finally {
			loading = false;
		}
	}

	function handleCloseReader() {
		activeChapter = null;
		view = 'detail';
	}

	async function handleUpload(file: File) {
		loading = true;
		try {
			const document = await api.uploadBook(file);
			if (document) {
				activeDocument = document;
				if (view !== 'detail' && view !== 'reader') {
					originNav = { view, plugin: currentPlugin, tabId: currentTabId };
				}
				view = 'detail';
				await refreshDocuments();
			}
		} finally {
			loading = false;
		}
	}
</script>

<div class="lector-app">
	{#if view !== 'reader'}
		<header class="mobile-header">
			<button class="menu-btn" onclick={() => (sidebarOpen = !sidebarOpen)}>
				{#if sidebarOpen}
					<X size={24} />
				{:else}
					<Menu size={24} />
				{/if}
			</button>
			<span class="app-name">Lector</span>
			<div class="header-spacer"></div>
		</header>

		<div class={clsx('sidebar-wrapper', sidebarOpen && 'open')}>
			<button
				class="sidebar-overlay"
				onclick={() => (sidebarOpen = false)}
				aria-label="Close sidebar"
			></button>
			<Sidebar {plugins} currentView={view} {currentTabId} onNavigate={handleNavigate} />
		</div>
	{/if}

	<main
		class={clsx(
			'main-viewport',
			view === 'reader' && 'full',
			view !== 'reader' && 'with-mobile-header'
		)}
	>
		{#key view}
			{#if view === 'library'}
				<Library
					{documents}
					{groups}
					onOpenDocument={(n) => handleSelectDocument(n.url, n.source)}
					onCreateGroup={handleCreateGroup}
					onUpload={handleUpload}
					onRefresh={refreshDocuments}
					onBatchDelete={handleBatchDelete}
					onBatchMove={handleBatchMove}
					onBatchArchive={handleBatchArchive}
					onBatchMarkRead={handleBatchMarkRead}
					onBatchRefresh={handleBatchRefresh}
				/>
			{:else if view === 'history'}
				<History
					{history}
					onOpenDocument={(n) => handleSelectDocument(n.url, n.source)}
					onRemove={async (id) => {
						await api.deleteHistory(id);
						await refreshHistory();
					}}
					onClearAll={async () => {
						confirmTitle = 'Clear History';
						confirmMessage = 'Are you sure you want to clear your entire reading history?';
						onConfirm = async () => {
							await api.clearHistory();
							await refreshHistory();
							showConfirmModal = false;
						};
						showConfirmModal = true;
					}}
					onBatchRemove={async (ids) => {
						await api.batchDeleteHistory(ids);
						await refreshHistory();
					}}
				/>
			{:else if view === 'search'}
				<Search
					{plugins}
					results={searchItems}
					{loading}
					bind:query={searchQuery}
					bind:source={searchSource}
					onSearch={handleSearch}
					onSelect={handleSelectDocument}
				/>
			{:else if view === 'plugins'}
				<PluginsView
					plugins={plugins.filter((p) => p.name !== 'system')}
					onRefresh={refreshPlugins}
				/>
			{:else if view === 'explorer'}
				<Explorer
					pluginName={currentPlugin}
					tabId={currentTabId}
					onSelectDocument={handleSelectDocument}
				/>
			{:else if view === 'dynamic'}
				<DynamicView pluginName={currentPlugin} tabId={currentTabId} />
			{:else if view === 'plugin'}
				<PluginExplorer
					name={currentPlugin}
					popular={popularResults}
					latest={latestResults}
					{loading}
					onSelect={handleSelectDocument}
				/>
			{:else if view === 'detail' && activeDocument}
				<Detail
					bind:document={activeDocument}
					{sources}
					onToggleLibrary={handleToggleLibrary}
					onReadChapter={handleReadChapter}
					onClose={() => handleNavigate(originNav.view, originNav.plugin, originNav.tabId)}
					backLabel={originNav.view === 'library'
						? 'Back to collection'
						: originNav.plugin !== 'system'
							? `Back to ${originNav.plugin}`
							: `Back to ${originNav.view}`}
					pluginManifests={plugins}
				/>
			{:else if view === 'settings'}
				<Settings pluginManifests={plugins} />
			{:else if view === 'reader' && activeChapter && activeDocument}
				<Reader
					chapter={activeChapter}
					document={activeDocument}
					onReadChapter={handleReadChapter}
					onClose={handleCloseReader}
					pluginManifests={plugins}
				/>
			{:else}
				<div class="empty-view">
					<p>Select a workspace item to begin</p>
				</div>
			{/if}
		{/key}
	</main>

	{#if loading}
		<div class="global-loader">
			<div class="loader-content">
				<span class="pulse"></span>
				<span>Syncing</span>
			</div>
		</div>
	{/if}

	<Modal
		show={showCreateGroupModal}
		title="Create New Group"
		onClose={() => (showCreateGroupModal = false)}
	>
		<div class="modal-form">
			<p>Organize your documents into groups.</p>
			<input
				type="text"
				bind:value={newGroupName}
				placeholder="Group name..."
				onkeydown={(e) => e.key === 'Enter' && submitCreateGroup()}
			/>
			<div class="modal-actions">
				<button class="modal-btn secondary" onclick={() => (showCreateGroupModal = false)}
					>Cancel</button
				>
				<button class="modal-btn primary" onclick={submitCreateGroup}>Create Group</button>
			</div>
		</div>
	</Modal>

	<Modal show={showConfirmModal} title={confirmTitle} onClose={() => (showConfirmModal = false)}>
		<div class="modal-confirm">
			<p>{confirmMessage}</p>
			<div class="modal-actions">
				<button class="modal-btn secondary" onclick={() => (showConfirmModal = false)}
					>Cancel</button
				>
				<button class="modal-btn danger-glow" onclick={() => onConfirm?.()}>Confirm</button>
			</div>
		</div>
	</Modal>
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
		background: var(--bg-main);
		overflow: hidden;
	}
	.lector-app {
		display: flex;
		height: 100vh;
		background: var(--bg-main);
		color: var(--text-main);
		position: relative;
	}
	.sidebar-wrapper {
		z-index: 1000;
	}
	.main-viewport {
		flex: 1;
		overflow-y: auto;
		padding: 3rem;
		background: var(--bg-main);
		transition: padding 0.3s ease;
	}
	.main-viewport.full {
		padding: 0;
	}
	.mobile-header {
		display: none;
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 60px;
		background: rgba(9, 9, 11, 0.8);
		backdrop-filter: blur(12px);
		border-bottom: 1px solid var(--border-main);
		z-index: 900;
		align-items: center;
		padding: 0 1rem;
	}
	.menu-btn {
		background: none;
		border: none;
		color: var(--text-main);
		cursor: pointer;
		padding: 0.5rem;
	}
	.app-name {
		flex: 1;
		text-align: center;
		font-weight: 700;
		letter-spacing: -0.01em;
		font-size: 1rem;
	}
	.header-spacer {
		width: 40px;
	}

	@media (max-width: 900px) {
		.mobile-header {
			display: flex;
		}
		.main-viewport {
			padding: 1.5rem;
		}
		.main-viewport.with-mobile-header {
			padding-top: calc(60px + 1.5rem);
		}
		.sidebar-wrapper {
			position: fixed;
			inset: 0;
			visibility: hidden;
			transition: visibility 0.3s;
		}
		.sidebar-wrapper.open {
			visibility: visible;
		}
		.sidebar-overlay {
			position: absolute;
			inset: 0;
			background: rgba(0, 0, 0, 0.6);
			backdrop-filter: blur(4px);
			opacity: 0;
			transition: opacity 0.3s;
			border: none;
			width: 100%;
			height: 100%;
		}
		.sidebar-wrapper.open .sidebar-overlay {
			opacity: 1;
		}
		:global(.sidebar) {
			position: absolute;
			left: -260px;
			top: 0;
			bottom: 0;
			transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		}
		.sidebar-wrapper.open :global(.sidebar) {
			transform: translateX(260px);
		}
	}

	.global-loader {
		position: fixed;
		bottom: 2rem;
		right: 2rem;
		z-index: 2000;
	}
	.loader-content {
		background: var(--text-main);
		color: var(--bg-main);
		padding: 0.6rem 1.2rem;
		font-size: 0.75rem;
		font-weight: 700;
		border-radius: 50px;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		box-shadow: 0 10px 20px rgba(0, 0, 0, 0.4);
	}
	.pulse {
		width: 8px;
		height: 8px;
		background: var(--primary);
		border-radius: 50%;
		animation: pulse 1.5s ease-in-out infinite;
	}
	@keyframes pulse {
		0% {
			transform: scale(0.8);
			opacity: 0.5;
		}
		50% {
			transform: scale(1.2);
			opacity: 1;
		}
		100% {
			transform: scale(0.8);
			opacity: 0.5;
		}
	}

	.modal-confirm p {
		color: var(--text-muted);
		margin-bottom: 1.5rem;
		font-size: 0.9rem;
	}
	.modal-form input {
		width: 100%;
		background: var(--bg-main);
		border: 1px solid var(--border-bright);
		color: var(--text-main);
		padding: 0.75rem 1rem;
		border-radius: 10px;
		margin-bottom: 2rem;
		font-size: 1rem;
		outline: none;
	}
	.modal-form input:focus {
		border-color: var(--primary);
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

	.modal-btn.primary {
		background: white;
		color: black;
		box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
	}

	.modal-btn.primary:hover {
		background: #f4f4f5;
		transform: translateY(-1px);
		box-shadow: 0 0 20px rgba(255, 255, 255, 0.2);
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
</style>
