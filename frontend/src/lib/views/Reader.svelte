<script lang="ts">
	import { onMount, tick } from 'svelte';
	import RotateCcw from 'lucide-svelte/icons/rotate-ccw';
	import CloudCheck from 'lucide-svelte/icons/cloud-check';
	import { clsx } from 'clsx';
	import { api, type PluginManifest } from '$lib/services/api';
	import ProgressRing from '../components/ProgressRing.svelte';
	import ReaderNav from '../components/reader/ReaderNav.svelte';
	import ReaderSettings from '../components/reader/ReaderSettings.svelte';
	import ReaderSidebar from '../components/reader/ReaderSidebar.svelte';
	import ReaderNotebook from '../components/reader/ReaderNotebook.svelte';
	import ReaderFooter from '../components/reader/ReaderFooter.svelte';
	import { toast } from '$lib/services/toast.svelte';
	import X from 'lucide-svelte/icons/x';
	import Loader2 from 'lucide-svelte/icons/loader-2';
	interface Chapter {
		id: number;
		title: string;
		order: number;
		content?: string;
		metadata?: string;
	}

	interface DocumentMeta {
		id: number;
		title: string;
		type: 'text' | 'images' | 'stream';
		source: string;
		chapters: { id: number; title: string; order: number }[];
	}

	interface Bookmark {
		id: number;
		chapter_id: number;
		title: string;
		created_at: string;
	}

	interface Note {
		id: number;
		content: string;
		created_at: string;
	}

	interface Props {
		chapter: Chapter;
		document: DocumentMeta;
		onClose: () => void;
		onReadChapter: (chapter: { id: number; title: string; order: number }) => void;
		pluginManifests?: PluginManifest[];
	}
	let { chapter, document: doc, onClose, onReadChapter, pluginManifests = [] }: Props = $props();

	let plugin = $derived(pluginManifests.find((p) => p.name === doc?.source));
	let readerOverride = $derived.by(() => {
		if (doc?.type) {
			const globalPlugin = pluginManifests.find((p) => p.ui_overrides?.[`reader:${doc.type}`]);
			if (globalPlugin) return globalPlugin.ui_overrides[`reader:${doc.type}`];
		}
		return plugin?.ui_overrides?.reader;
	});

	let settings = $state({
		fontSize: 18,
		fontFamily: 'serif',
		theme: 'slate' as 'slate' | 'sepia' | 'light',
		lineHeight: 1.6,
		paragraphSpacing: 1.5,
		sideMargin: 12,
		topMargin: 60,
		readingMode: 'paged' as 'paged' | 'infinite'
	});

	let showUI = $state(false);
	let showSettings = $state(false);
	let showSidebar = $state(false);
	let showBookmarks = $state(false);
	let showNotes = $state(false);
	let isSynced = $state(false);

	let chaptersStack = $state<Chapter[]>([]);
	let isLoadingNext = $state(false);
	let currentInViewId = $state<number>(0);
	let currentChapterScroll = $state(0);
	let bookmarks = $state<Bookmark[]>([]);
	let notes = $state<Note[]>([]);
	let newNoteText = $state('');

	let isRestoring = $state(true);
	let lastSyncedId = 0;
	let lastSyncedPos = -1;

	let isSpeaking = $state(false);
	let synth: SpeechSynthesis | null = null;
	let utterance: SpeechSynthesisUtterance | null = null;

	let selectedText = $state('');
	let selectionCoords = $state({ x: 0, y: 0 });
	let actionResult = $state<any>(null);
	let actionLoading = $state(false);

	const selectionActions = $derived.by(() => {
		const actions = pluginManifests.flatMap((p) =>
			(p.actions || [])
				.filter((a: any) => a.context === 'selection')
				.map((a: any) => ({ ...a, plugin: p.name }))
		);
		return actions;
	});

	function handleSelectionChange() {
		const selection = window.getSelection();
		if (!selection || selection.isCollapsed || !selection.toString().trim()) {
			selectedText = '';
			return;
		}

		const text = selection.toString().trim();
		if (text.length > 100) {
			selectedText = '';
			return;
		}

		const range = selection.getRangeAt(0);
		const rect = range.getBoundingClientRect();

		selectedText = text;
		selectionCoords = {
			x: rect.left + rect.width / 2,
			y: rect.top
		};
	}

	async function handleAction(pluginName: string, method: string) {
		if (!selectedText) return;
		const word = selectedText;
		selectedText = '';

		actionLoading = true;
		actionResult = null;
		try {
			const res = await fetch(`${window.location.origin}/api/plugins/${pluginName}/rpc/${method}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(word)
			});
			if (!res.ok) throw new Error(await res.text());
			const data = await res.json();
			if (data.error) {
				toast.error(data.error);
			} else {
				actionResult = data[0] || data;
			}
		} catch (e: any) {
			toast.error(e.message || 'Action failed');
		} finally {
			actionLoading = false;
		}
	}

	$effect(() => {
		if (
			chapter &&
			(chaptersStack.length === 0 || !chaptersStack.some((c) => c.id === chapter.id))
		) {
			chaptersStack = [chapter];
			currentInViewId = chapter.id;
			const container = document.querySelector('.main-viewport');
			if (container && !isRestoring) container.scrollTo({ top: 0, behavior: 'instant' });
			if (!isRestoring) syncProgress(chapter.id, 0);
		}
	});

	async function syncProgress(idOverride?: number, posOverride?: number) {
		if (isRestoring) return;

		const id = idOverride || currentInViewId;
		if (!id || !doc?.id) return;

		const container = document.querySelector('.main-viewport');
		if (!container) return;

		let scrollPos = posOverride !== undefined ? posOverride : 0;
		if (posOverride === undefined) {
			const block = container.querySelector(`[data-ch-id="${id}"]`) as HTMLElement;
			if (block) {
				const containerRect = container.getBoundingClientRect();
				const blockRect = block.getBoundingClientRect();
				scrollPos = (containerRect.top - blockRect.top) / blockRect.height;
			} else {
				scrollPos = container.scrollTop / (container.scrollHeight - container.clientHeight);
			}
		}

		scrollPos = Math.max(0, Math.min(1, scrollPos));
		currentChapterScroll = scrollPos * 100;

		if (id === lastSyncedId && Math.abs(scrollPos - lastSyncedPos) < 0.005) return;

		lastSyncedId = id;
		lastSyncedPos = scrollPos;

		try {
			await fetch('/api/progress', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				keepalive: true,
				body: JSON.stringify({
					document_id: doc.id,
					chapter_id: id,
					scroll_pos: scrollPos,
					client_updated_at: Date.now()
				})
			});
			isSynced = true;
			setTimeout(() => (isSynced = false), 2000);
		} catch {}
	}

	function handleScroll() {
		const container = document.querySelector('.main-viewport');
		if (!container || isRestoring) return;

		const centerY = container.getBoundingClientRect().top + container.clientHeight / 2;
		const blocks = Array.from(container.querySelectorAll('.chapter-block'));

		let activeBlock: HTMLElement | null = null;
		for (const block of blocks) {
			const rect = block.getBoundingClientRect();
			if (centerY >= rect.top && centerY <= rect.bottom) {
				activeBlock = block as HTMLElement;
				break;
			}
		}

		if (activeBlock) {
			const id = parseInt(activeBlock.getAttribute('data-ch-id') || '0');
			if (id && id !== currentInViewId) {
				currentInViewId = id;
				syncProgress(id, 0);
			}

			const containerRect = container.getBoundingClientRect();
			const blockRect = activeBlock.getBoundingClientRect();
			const progress = (containerRect.top - blockRect.top) / blockRect.height;
			currentChapterScroll = Math.max(0, Math.min(100, progress * 100));
		}

		if (settings.readingMode === 'infinite' && !isLoadingNext) {
			const threshold = 1800;
			if (container.scrollHeight - container.scrollTop - container.clientHeight < threshold) {
				loadNextInStack();
			}
		}
	}

	onMount(() => {
		const handleExitSync = () => syncProgress();
		const container = document.querySelector('.main-viewport');

		window.document.addEventListener('selectionchange', handleSelectionChange);

		const init = async () => {
			const saved = localStorage.getItem('lector-reader-settings');
			if (saved) {
				const parsed = JSON.parse(saved);
				if (parsed.readingMode === 'tap') parsed.readingMode = 'paged';
				settings = { ...settings, ...parsed };
			}

			synth = window.speechSynthesis;
			if (container) container.addEventListener('scroll', handleScroll);

			window.addEventListener('pagehide', handleExitSync);
			window.addEventListener('beforeunload', handleExitSync);

			if (doc?.id) {
				await Promise.all([refreshBookmarks(), refreshNotes()]);
				const prog = await api.getDocumentProgress(doc.id);

				if (prog && prog.chapter_id === chapter.id && container) {
					const performRestore = () => {
						const block = container.querySelector(
							`[data-ch-id="${prog.chapter_id}"]`
						) as HTMLElement;
						if (block) {
							const targetY = block.offsetTop + prog.scroll_pos * block.offsetHeight;
							container.scrollTo({ top: targetY, behavior: 'instant' });
							return true;
						}
						return false;
					};

					let attempts = 0;
					const restoreLoop = () => {
						attempts++;
						const success = performRestore();
						if (!success && attempts < 15) {
							requestAnimationFrame(restoreLoop);
						} else {
							setTimeout(() => {
								isRestoring = false;
							}, 200);
						}
					};
					requestAnimationFrame(restoreLoop);
				} else {
					isRestoring = false;
				}
			} else {
				isRestoring = false;
			}
		};

		init();
		const progressInterval = setInterval(() => syncProgress(), 5000);
		const analyticsInterval = setInterval(() => {
			if (document.visibilityState === 'visible' && !isRestoring) {
				api.trackAnalytics('time', 60);
			}
		}, 60000);

		return () => {
			clearInterval(progressInterval);
			clearInterval(analyticsInterval);
			window.document.removeEventListener('selectionchange', handleSelectionChange);
			if (container) container.removeEventListener('scroll', handleScroll);
			window.removeEventListener('pagehide', handleExitSync);
			window.removeEventListener('beforeunload', handleExitSync);
			syncProgress();
			stopTTS();
		};
	});

	async function refreshBookmarks() {
		bookmarks = await api.getBookmarks(doc.id);
	}
	async function refreshNotes() {
		notes = await api.getNotes(doc.id);
	}
	async function handleAddBookmark() {
		const ch = doc.chapters.find((c) => c.id === currentInViewId);
		await api.addBookmark(doc.id, currentInViewId, ch?.title || 'Bookmark');
		await refreshBookmarks();
	}
	async function handleDeleteBookmark(id: number) {
		await api.deleteBookmark(id);
		await refreshBookmarks();
	}
	async function handleAddNote() {
		if (!newNoteText.trim()) return;
		await api.addNote(doc.id, currentInViewId, newNoteText, '');
		newNoteText = '';
		await refreshNotes();
	}
	async function handleDeleteNote(id: number) {
		await api.deleteNote(id);
		await refreshNotes();
	}

	async function loadNextInStack() {
		const lastCh = chaptersStack[chaptersStack.length - 1];
		const currentIdx = doc.chapters.findIndex((c) => c.id === lastCh.id);
		const nextMeta = doc.chapters[currentIdx + 1];
		if (nextMeta && !isLoadingNext) {
			isLoadingNext = true;
			try {
				const fullNext = await api.getChapter(nextMeta.id);
				chaptersStack = [...chaptersStack, fullNext];
				await tick();
			} finally {
				isLoadingNext = false;
			}
		}
	}

	$effect(() => {
		localStorage.setItem('lector-reader-settings', JSON.stringify(settings));
	});

	function navigate(offset: number) {
		stopTTS();
		const currentIdx = doc.chapters.findIndex((c) => c.id === currentInViewId);
		const target = doc.chapters[currentIdx + offset];
		if (target) {
			onReadChapter(target);
			syncProgress(target.id, 0);
		}
	}

	function toggleTTS() {
		if (!synth) return;
		if (isSpeaking) {
			stopTTS();
		} else {
			const text = document.querySelector('.prose')?.textContent || '';
			utterance = new SpeechSynthesisUtterance(text);
			utterance.onend = () => (isSpeaking = false);
			synth.speak(utterance);
			isSpeaking = true;
		}
	}
	function stopTTS() {
		if (synth) {
			synth.cancel();
			isSpeaking = false;
		}
	}

	function handleReaderClick(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (target.closest('button, select, input, .reader-sidebar, .settings-panel, textarea')) return;
		showUI = !showUI;
	}

	const currentIdxMeta = $derived(doc.chapters.findIndex((c) => c.id === currentInViewId));
	const hasPrev = $derived(currentIdxMeta > 0);
	const hasNext = $derived(currentIdxMeta < doc.chapters.length - 1);

	function parseMetadata(metadata?: string) {
		try {
			return JSON.parse(metadata || '');
		} catch {
			return null;
		}
	}
</script>

{#snippet textView(ch: Chapter)}
	<article class="prose">
		{#if ch.content}
			<!-- eslint-disable-next-line svelte/no-at-html-tags -->
			{@html ch.content}
		{:else}
			<div class="sync-error"><h3>Syncing...</h3></div>
		{/if}
	</article>
{/snippet}

{#snippet imagesView(ch: Chapter)}
	<div class="images-viewer">
		{#each parseMetadata(ch.metadata) || [] as img, i (i)}
			<img
				src={api.getProxyImage(img)}
				alt="Page {i + 1}"
				class="images-page"
				loading="lazy"
				onerror={(e) => ((e.currentTarget as HTMLElement).style.display = 'none')}
			/>
		{/each}
	</div>
{/snippet}

{#if chapter}
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		class={clsx('reader-container', `theme-${settings.theme}`, `font-${settings.fontFamily}`)}
		onclick={handleReaderClick}
		onmouseenter={() => (showUI = true)}
		onmouseleave={() => {
			showUI = false;
			showSettings = false;
		}}
	>
		<button
			type="button"
			class="ui-trigger-zone"
			onclick={() => (showUI = !showUI)}
			aria-label="Toggle UI"
		></button>

		<ReaderNav
			{showUI}
			docTitle={doc.title}
			{isSpeaking}
			onToggleSidebar={() => (showSidebar = !showSidebar)}
			onToggleBookmarks={() => (showBookmarks = !showBookmarks)}
			onToggleSettings={() => (showSettings = !showSettings)}
			onToggleNotes={() => (showNotes = !showNotes)}
			onToggleTTS={toggleTTS}
			{onClose}
		/>

		{#if showSettings}
			<ReaderSettings bind:settings />
		{/if}

		{#if showSidebar || showBookmarks}
			<ReaderSidebar
				{showSidebar}
				{showBookmarks}
				chapters={doc.chapters}
				{bookmarks}
				currentChapterId={currentInViewId}
				onReadChapter={(ch) => {
					onReadChapter(ch);
					showSidebar = false;
				}}
				onAddBookmark={handleAddBookmark}
				onDeleteBookmark={handleDeleteBookmark}
				onClose={() => {
					showSidebar = false;
					showBookmarks = false;
				}}
			/>
		{/if}

		{#if showNotes}
			<ReaderNotebook
				{notes}
				bind:newNoteText
				onAddNote={handleAddNote}
				onDeleteNote={handleDeleteNote}
				onClose={() => (showNotes = false)}
			/>
		{/if}

		<main
			class={clsx(
				'reader-main',
				readerOverride && readerOverride.type === 'iframe' && 'is-override'
			)}
			style="--font-size: {settings.fontSize}px; --line-height: {settings.lineHeight}; --p-spacing: {settings.paragraphSpacing}em; --side-margin: {settings.sideMargin}%; --top-margin: {settings.topMargin}px;"
		>
			{#if readerOverride && readerOverride.type === 'iframe'}
				<iframe
					src={`${readerOverride.url}?doc_id=${doc.id}&ch_id=${currentInViewId || chapter.id}`}
					style="position: fixed; top: 0; left: 0; width: 100vw; height: 100vh; border: none; z-index: 100; background: #000;"
					title={chapter.title}
					allowfullscreen
				></iframe>
			{:else}
				{#each chaptersStack as ch (ch.id)}
					<div class={clsx('chapter-block', doc.type)} data-ch-id={ch.id}>
						{#if doc.type !== 'stream'}
							<header class="chapter-header"><h1>{ch.title}</h1></header>
						{/if}

						{#if doc.type === 'images'}
							{@render imagesView(ch)}
						{:else}
							{@render textView(ch)}
						{/if}

						{#if settings.readingMode === 'infinite'}<div class="stack-divider"></div>{/if}
					</div>
				{/each}

				{#if settings.readingMode !== 'infinite'}
					<ReaderFooter {hasPrev} {hasNext} currentIdx={currentIdxMeta} onNavigate={navigate} />
				{/if}

				{#if isLoadingNext}
					<div class="stack-loader">
						<RotateCcw size={20} class="spin" /><span>Seamlessly loading next chapter...</span>
					</div>
				{/if}
			{/if}
		</main>

		<div class="progress-badge">
			{#if isSynced}
				<CloudCheck size={14} class="sync-icon" />
			{/if}
			<ProgressRing value={currentChapterScroll} total={100} size={20} stroke={2} />
			<span>{currentIdxMeta + 1} / {doc.chapters.length}</span>
		</div>

		{#if selectedText && selectionActions.length > 0}
			<div
				class="selection-toolbar"
				style="left: {selectionCoords.x}px; top: {selectionCoords.y}px"
			>
				{#each selectionActions as action (action.label)}
					<button class="tool-btn" onclick={() => handleAction(action.plugin, action.method)}>
						<span>{action.label}</span>
					</button>
				{/each}
			</div>
		{/if}

		{#if actionResult || actionLoading}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div class="definition-overlay" onclick={() => (actionResult = null)}>
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div class="definition-card" onclick={(e) => e.stopPropagation()}>
					{#if actionLoading}
						<div class="def-loader">
							<Loader2 size={24} class="spin" />
							<span>Processing action...</span>
						</div>
					{:else if actionResult}
						<header>
							<div class="word-meta">
								<h2>{actionResult.word || actionResult.title || 'Result'}</h2>
								{#if actionResult.phonetic}
									<span class="phonetic">{actionResult.phonetic}</span>
								{/if}
							</div>
							<button class="close-btn" onclick={() => (actionResult = null)}>
								<X size={18} />
							</button>
						</header>
						<div class="def-body">
							{#if actionResult.meanings}
								{#each actionResult.meanings as meaning, i (i)}
									<div class="meaning">
										<span class="part-of-speech">{meaning.partOfSpeech}</span>
										<ul class="definitions">
											{#each meaning.definitions.slice(0, 2) as def, j (j)}
												<li>{def.definition}</li>
											{/each}
										</ul>
									</div>
								{/each}
							{:else if actionResult.message}
								<p class="raw-result">{actionResult.message}</p>
							{:else}
								<p class="raw-result">{JSON.stringify(actionResult, null, 2)}</p>
							{/if}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.reader-container {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		position: relative;
		transition: background 0.3s ease;
		overflow-x: hidden;
		overflow-anchor: none;
	}
	button {
		appearance: none;
		background: none;
		border: none;
		padding: 0;
		margin: 0;
		color: inherit;
		font-family: inherit;
		cursor: pointer;
		outline: none;
	}
	.theme-slate {
		--bg: var(--bg-main);
		--text: var(--text-muted);
		--text-strong: var(--text-main);
		--surface: var(--bg-surface);
	}
	.theme-sepia {
		--bg: #f4ecd8;
		--text: #5b4636;
		--text-strong: #1a1a1a;
		--surface: #e8dfc4;
		--border-main: #d3cbb0;
	}
	.theme-light {
		--bg: #ffffff;
		--text: #333333;
		--text-strong: #000000;
		--surface: #f5f5f5;
		--border-main: #e5e5e5;
	}
	.reader-container {
		background: var(--bg);
		color: var(--text);
	}
	.font-serif {
		font-family: serif;
	}
	.font-sans {
		font-family: var(--font-sans);
	}
	.font-mono {
		font-family: var(--font-mono);
	}
	.font-system {
		font-family: system-ui;
	}

	.reader-main {
		max-width: 900px;
		margin: 0 auto;
		padding: var(--top-margin) var(--side-margin) 10rem;
		font-size: var(--font-size);
		line-height: var(--line-height);
	}
	:global(.reader-main.is-override) {
		max-width: none !important;
		width: 100% !important;
		padding: 0 !important;
		margin: 0 !important;
		flex: 1 !important;
		height: 100vh !important;
		display: flex !important;
		flex-direction: column !important;
	}
	:global(.stream) .reader-main {
		max-width: 1200px;
	}
	:global(.images) .reader-main {
		max-width: 1000px;
		padding-left: 0;
		padding-right: 0;
	}
	.chapter-header {
		text-align: center;
		margin-bottom: 3rem;
		color: var(--text-strong);
	}
	h1 {
		font-size: 2.2em;
		font-weight: 700;
		letter-spacing: -0.02em;
	}
	:global(.prose p) {
		margin-bottom: var(--p-spacing);
	}
	:global(.prose img.comic-page) {
		max-width: 100%;
		height: auto;
		display: block;
		margin: 0 auto 1.5rem auto;
		border-radius: 12px;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
	}

	.images-viewer {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0;
	}
	.images-page {
		max-width: 100%;
		height: auto;
		display: block;
	}

	.stream-viewer {
		width: 100%;
		aspect-ratio: 16 / 9;
		background: black;
		border-radius: 12px;
		overflow: hidden;
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
	}
	.stream-element,
	.stream-frame {
		width: 100%;
		height: 100%;
		display: block;
	}

	.stack-divider {
		height: 1px;
		background: var(--border-main);
		margin: 6rem auto;
		width: 40%;
		opacity: 0.5;
	}
	.stack-loader {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 4rem 0;
		color: var(--text-strong);
		font-weight: 600;
		font-size: 0.875rem;
	}
	.spin {
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

	.ui-trigger-zone {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		height: 120px;
		z-index: 100;
		background: transparent;
		border: none;
		cursor: default;
	}
	.progress-badge {
		position: fixed;
		bottom: 1.5rem;
		right: 1.5rem;
		background: var(--surface);
		border: 1px solid var(--border-main);
		padding: 0.5rem 0.8rem;
		border-radius: 20px;
		font-size: 0.65rem;
		font-weight: 700;
		z-index: 50;
		color: var(--text-strong);
		display: flex;
		align-items: center;
		gap: 0.75rem;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}
	.sync-icon {
		color: #10b981;
		animation: fadeIn 0.3s ease;
	}
	.selection-toolbar {
		position: fixed;
		transform: translate(-50%, -120%);
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		padding: 0.4rem;
		border-radius: 10px;
		display: flex;
		gap: 0.5rem;
		z-index: 9999;
		box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
		animation: flyUp 0.2s cubic-bezier(0.4, 0, 0.2, 1);
		pointer-events: auto;
	}
	.tool-btn {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--text-main);
		white-space: nowrap;
	}
	.tool-btn:hover {
		background: rgba(255, 255, 255, 0.05);
		color: var(--primary);
	}
	.definition-overlay {
		position: fixed;
		inset: 0;
		background: rgba(0, 0, 0, 0.4);
		backdrop-filter: blur(4px);
		z-index: 2000;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}
	.definition-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 20px;
		width: 100%;
		max-width: 440px;
		max-height: 80vh;
		display: flex;
		flex-direction: column;
		box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
		animation: scaleUp 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
		overflow: hidden;
	}
	.definition-card header {
		padding: 1.5rem;
		border-bottom: 1px solid var(--border-main);
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}
	.word-meta h2 {
		margin: 0;
		font-size: 1.5rem;
		font-weight: 800;
		color: var(--text-main);
		text-transform: capitalize;
	}
	.phonetic {
		font-size: 0.875rem;
		color: var(--primary);
		font-family: var(--font-mono);
		margin-top: 0.25rem;
		display: block;
	}
	.def-body {
		padding: 1.5rem;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}
	.meaning {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}
	.part-of-speech {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-dim);
		background: var(--bg-main);
		padding: 0.2rem 0.6rem;
		border-radius: 4px;
		width: fit-content;
	}
	.definitions {
		margin: 0;
		padding-left: 1.25rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	.definitions li {
		font-size: 0.9375rem;
		line-height: 1.5;
		color: var(--text-main);
	}
	.def-loader {
		padding: 4rem 2rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		color: var(--text-dim);
	}
	@keyframes flyUp {
		from {
			opacity: 0;
			transform: translate(-50%, -100%);
		}
		to {
			opacity: 1;
			transform: translate(-50%, -120%);
		}
	}
	@keyframes scaleUp {
		from {
			opacity: 0;
			transform: scale(0.95);
		}
		to {
			opacity: 1;
			transform: scale(1);
		}
	}
	@media (max-width: 600px) {
		.reader-main {
			padding: 40px 1rem 10rem;
		}
	}
</style>
