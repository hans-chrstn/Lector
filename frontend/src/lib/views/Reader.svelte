<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { RotateCcw, CloudCheck } from 'lucide-svelte';
	import { clsx } from 'clsx';
	import { api } from '$lib/services/api';
	import ProgressRing from '../components/ProgressRing.svelte';
	import ReaderNav from '../components/reader/ReaderNav.svelte';
	import ReaderSettings from '../components/reader/ReaderSettings.svelte';
	import ReaderSidebar from '../components/reader/ReaderSidebar.svelte';
	import ReaderNotebook from '../components/reader/ReaderNotebook.svelte';
	import ReaderFooter from '../components/reader/ReaderFooter.svelte';

	interface Chapter {
		id: number;
		title: string;
		order: number;
		content?: string;
	}

	interface DocumentMeta {
		id: number;
		title: string;
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
	}
	let { chapter, document: doc, onClose, onReadChapter }: Props = $props();

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

		return () => {
			clearInterval(progressInterval);
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
</script>

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
			class="reader-main"
			style="--font-size: {settings.fontSize}px; --line-height: {settings.lineHeight}; --p-spacing: {settings.paragraphSpacing}em; --side-margin: {settings.sideMargin}%; --top-margin: {settings.topMargin}px;"
		>
			{#each chaptersStack as ch (ch.id)}
				<div class="chapter-block" data-ch-id={ch.id}>
					<header class="chapter-header"><h1>{ch.title}</h1></header>
					<article class="prose">
						{#if ch.content}
							<!-- eslint-disable-next-line svelte/no-at-html-tags -->
							{@html ch.content}
						{:else}
							<div class="sync-error"><h3>Syncing...</h3></div>
						{/if}
					</article>
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
		</main>

		<div class="progress-badge">
			{#if isSynced}
				<CloudCheck size={14} class="sync-icon" />
			{/if}
			<ProgressRing value={currentChapterScroll} total={100} size={20} stroke={2} />
			<span>{currentIdxMeta + 1} / {doc.chapters.length}</span>
		</div>
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
	@media (max-width: 600px) {
		.reader-main {
			padding: 40px 1rem 10rem;
		}
	}
</style>
