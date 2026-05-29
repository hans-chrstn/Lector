<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { ArrowLeft } from 'lucide-svelte';
	import {
		api,
		type Document,
		type Chapter,
		type SearchItem,
		type PluginManifest
	} from '$lib/services/api';
	import { SvelteMap } from 'svelte/reactivity';

	import DetailHero from './detail/DetailHero.svelte';
	import DetailChapters from './detail/DetailChapters.svelte';
	import DetailModals from './detail/DetailModals.svelte';

	interface Props {
		document: Document;
		sources?: string[];
		onToggleLibrary: () => void;
		onReadChapter: (chapter: Chapter) => void;
		onClose: () => void;
		backLabel?: string;
		pluginManifests?: PluginManifest[];
	}

	let {
		document = $bindable(),
		sources = [],
		onToggleLibrary,
		onReadChapter,
		onClose,
		backLabel = 'Back to collection',
		pluginManifests = []
	}: Props = $props();

	function capitalize(s: string) {
		return s.charAt(0).toUpperCase() + s.slice(1);
	}

	let dynamicActions = $state<any[]>([]);
	let progress = $state<any>(null);
	let pollingInterval: any;
	let pollCount = 0;
	const MAX_POLLS = 10;

	let showEdit = $state(false);
	let showMigrate = $state(false);

	const isReady = $derived(document.chapters && document.chapters.length > 0);
	const hasStarted = $derived(progress && progress.chapter_id);

	const availableActions = $derived(() => {
		const manifestActions = pluginManifests.flatMap((p) =>
			(p.actions || [])
				.filter((a) => a.context === 'document_detail' && a.method !== 'open_metadata_modal')
				.map((a) => ({ ...a, plugin: p.name }))
		);
		const actionMap = new SvelteMap();
		for (const a of manifestActions) actionMap.set(a.label, a);
		for (const a of dynamicActions) actionMap.set(a.label, a);
		return Array.from(actionMap.values());
	});

	async function fetchDynamicActions() {
		if (!document.id) return;
		dynamicActions = [];
		for (const p of pluginManifests) {
			if (p.name === 'system' || !p.is_loaded) continue;
			try {
				const res = await fetch(
					`${window.location.origin}/api/plugins/${p.name}/rpc/get_document_actions`,
					{
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(document)
					}
				);
				if (res.ok) {
					const actions = await res.json();
					if (Array.isArray(actions)) {
						dynamicActions = [...dynamicActions, ...actions.map((a) => ({ ...a, plugin: p.name }))];
					}
				}
			} catch (e) {
				console.error(`[Detail] Failed to fetch actions for plugin ${p.name}:`, e);
			}
		}
	}

	async function fetchProgress() {
		try {
			const res = await api.getDocumentProgress(document.id);
			if (res) progress = res;
		} catch {}
	}

	async function handlePluginAction(pluginName: string, method: string) {
		try {
			const res = await fetch(`${window.location.origin}/api/plugins/${pluginName}/rpc/${method}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(document)
			});
			if (!res.ok) alert(await res.text());
			else {
				const data = await res.json();
				if (data.download_url) {
					const link = window.document.createElement('a');
					link.href = data.download_url;
					link.download = '';
					window.document.body.appendChild(link);
					link.click();
					window.document.body.removeChild(link);
				}
				if (data.message) alert(data.message);
				await fetchDynamicActions();
			}
		} catch (e) {
			console.error('Failed to execute plugin action:', e);
			alert('Failed to execute plugin action');
		}
	}

	async function handleReadAction() {
		if (!isReady || !document.chapters) return;
		await fetchProgress();
		let targetChapter = document.chapters[0];
		if (hasStarted) {
			const saved = document.chapters.find((c) => c.id === progress.chapter_id);
			if (saved) targetChapter = saved;
		}
		onReadChapter(targetChapter);
	}

	async function handleBatchAction(ids: number[], isRead: boolean) {
		await api.batchUpdateChapters(ids, isRead);
		document.chapters = document.chapters.map((c) => {
			if (ids.includes(c.id)) return { ...c, is_read: isRead };
			return c;
		});
		document.read_chapters = document.chapters.filter((c) => c.is_read).length;
	}

	async function handleSaveEdit(form: any) {
		await api.updateMetadata(document.id, form);
		document = { ...document, ...form };
		showEdit = false;
	}

	async function handleSelectMigrate(result: SearchItem, source: string) {
		await api.migrateDocument(document.id, result.url, source);
		document = await api.ensureDocument(result.url, source);
		showMigrate = false;
	}

	async function handleCoverUpload(file: File) {
		const res = await api.updateCover(document.id, file);
		document.cover_url = res.url;
	}

	function startPolling() {
		stopPolling();
		pollCount = 0;
		pollingInterval = setInterval(async () => {
			pollCount++;
			const updated = await api.getDocument(document.id);
			if (updated.chapters && updated.chapters.length > 0) {
				document = updated;
				stopPolling();
			}
			if (pollCount >= MAX_POLLS) stopPolling();
		}, 2000);
	}

	function stopPolling() {
		if (pollingInterval) clearInterval(pollingInterval);
		pollingInterval = null;
	}

	onMount(async () => {
		if (!isReady) startPolling();
		await Promise.all([fetchProgress(), fetchDynamicActions()]);
	});

	onDestroy(() => stopPolling());

	$effect(() => {
		if (document.id && pluginManifests.length > 0) fetchDynamicActions();
	});
</script>

<div class="detail-container">
	<button class="back-link" onclick={onClose}>
		<ArrowLeft size={18} />
		<span>{capitalize(backLabel)}</span>
	</button>

	<DetailHero
		bind:document
		{isReady}
		{hasStarted}
		availableActions={availableActions() as any[]}
		onReadAction={handleReadAction}
		{onToggleLibrary}
		onEdit={() => (showEdit = true)}
		onCoverUpload={handleCoverUpload}
		onPluginAction={handlePluginAction}
	/>

	<DetailChapters
		chapters={document.chapters}
		readChapters={document.read_chapters}
		{progress}
		{onReadChapter}
		onBatchAction={handleBatchAction}
	/>

	<DetailModals
		{document}
		{sources}
		{pluginManifests}
		{showEdit}
		onCloseEdit={() => (showEdit = false)}
		onSaveEdit={handleSaveEdit}
		{showMigrate}
		onCloseMigrate={() => (showMigrate = false)}
		onSelectMigrate={handleSelectMigrate}
	/>
</div>

<style>
	.detail-container {
		padding-bottom: 5rem;
		animation: fadeIn 0.3s ease-out;
		max-width: 1400px;
		margin: 0 auto;
	}

	.back-link {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-dim);
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		margin-bottom: 2.5rem;
		padding: 0.5rem 0;
		transition: color 0.2s;
	}

	.back-link:hover {
		color: var(--text-main);
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateY(10px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style>
