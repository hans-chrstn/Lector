<script lang="ts">
	import Play from 'lucide-svelte/icons/play';
	import Download from 'lucide-svelte/icons/download';
	import LibraryIcon from 'lucide-svelte/icons/library';
	import Edit3 from 'lucide-svelte/icons/edit-3';
	import User from 'lucide-svelte/icons/user';
	import BookOpen from 'lucide-svelte/icons/book-open';
	import Tag from 'lucide-svelte/icons/tag';
	import Zap from 'lucide-svelte/icons/zap';
	import Compass from 'lucide-svelte/icons/compass';
	import Settings from 'lucide-svelte/icons/settings';
	import CloudDownload from 'lucide-svelte/icons/cloud-download';
	import CloudUpload from 'lucide-svelte/icons/cloud-upload';
	import FileText from 'lucide-svelte/icons/file-text';
	import RefreshCw from 'lucide-svelte/icons/refresh-cw';
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import { clsx } from 'clsx';

	const IconMap: Record<string, any> = {
		Compass,
		Settings,
		CloudDownload,
		CloudUpload,
		FileText,
		RefreshCw,
		Trash2,
		Zap,
		Tag,
		User,
		BookOpen
	};
	import type { Document } from '$lib/services/api';
	import CoverImage from '../../components/CoverImage.svelte';

	interface Props {
		document: Document;
		isReady: boolean;
		hasStarted: boolean | null;
		availableActions: any[];
		onReadAction: () => void;
		onToggleLibrary: () => void;
		onEdit: () => void;
		onCoverUpload: (file: File) => void;
		onPluginAction: (plugin: string, method: string) => void;
		onRefresh: () => void;
	}

	let {
		document = $bindable(),
		isReady,
		hasStarted,
		availableActions,
		onReadAction,
		onToggleLibrary,
		onEdit,
		onCoverUpload,
		onPluginAction,
		onRefresh
	}: Props = $props();

	let coverInput: HTMLInputElement;
	let synopsisEl = $state<HTMLElement>();
	let isExpanded = $state(false);
	let isTruncated = $state(false);

	$effect(() => {
		if (synopsisEl && document.synopsis) {
			isTruncated = synopsisEl.scrollHeight > synopsisEl.clientHeight;
		}
	});
</script>

<div class="hero">
	<div class="cover-section">
		<button class="cover-wrapper" onclick={() => coverInput.click()} title="Change Cover">
			<CoverImage src={document.cover_url} alt={document.title} isHero={true} />
			<div class="cover-overlay">
				<Edit3 size={24} />
			</div>
		</button>
		<input
			type="file"
			accept="image/*"
			bind:this={coverInput}
			onchange={(e) => e.currentTarget.files && onCoverUpload(e.currentTarget.files[0])}
			hidden
		/>
	</div>

	<div class="info-section">
		<div class="top-meta">
			<span class="source-tag">{document.source}</span>
			{#if document.status}
				<span class="status-tag">{document.status}</span>
			{/if}
		</div>

		<h1>{document.title}</h1>

		<div class="meta-grid">
			<div class="meta-item">
				<User size={16} />
				<div class="meta-label">
					<span class="label">Author</span>
					<span class="value">{document.author || 'Unknown Author'}</span>
				</div>
			</div>
			<div class="meta-item">
				<BookOpen size={16} />
				<div class="meta-label">
					<span class="label">Volume</span>
					<span class="value"
						>{document.total_chapters || document.chapters?.length || 0} Chapters</span
					>
				</div>
			</div>
			<div class="meta-item">
				<Tag size={16} />
				<div class="meta-label">
					<span class="label">Genres</span>
					<span class="value genres">{document.genres || 'No genres listed'}</span>
				</div>
			</div>
		</div>

		{#if document.synopsis}
			<div class="synopsis-box">
				<h3>Synopsis</h3>
				<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<p
					bind:this={synopsisEl}
					class={clsx('synopsis', isExpanded && 'expanded', isTruncated && 'clickable')}
					onclick={() => isTruncated && (isExpanded = !isExpanded)}
				>
					{document.synopsis}
				</p>
				{#if isTruncated}
					<button class="expand-toggle" onclick={() => (isExpanded = !isExpanded)}>
						{isExpanded ? 'Show less' : 'Show more'}
					</button>
				{/if}
			</div>
		{/if}

		<div class="main-actions">
			{#if isReady}
				<button class="primary-btn" onclick={onReadAction}>
					<Play size={18} fill="currentColor" />
					<span
						>{hasStarted ? 'Continue' : 'Start'}
						{['stream', 'video'].includes(document.type?.toLowerCase() || '')
							? 'watching'
							: 'reading'}</span
					>
				</button>

				{#if document.is_in_library}
					<button class="secondary-btn" onclick={onRefresh}>
						<RefreshCw size={18} />
						<span>Refresh</span>
					</button>
					{#if !availableActions.some((a) => a.label === 'Export EPUB' || a.label === 'Downloaded' || a.label === 'Download Offline')}
						<a
							href={`${window.location.origin}/api/documents/${document.id}/export?format=epub`}
							class="secondary-btn"
							download={`${document.title}.epub`}
							rel="external"
						>
							<Download size={18} />
							<span>Export</span>
						</a>
					{/if}
				{/if}

				{#each availableActions as action, i (i)}
					{@const Icon = IconMap[action.icon] || Zap}
					<button
						class="secondary-btn"
						onclick={() => onPluginAction(action.plugin, action.method)}
					>
						<Icon size={18} />
						<span>{action.label}</span>
					</button>
				{/each}
			{/if}

			<button
				class={clsx('icon-action-btn', document.is_in_library && 'active')}
				onclick={onToggleLibrary}
				title={document.is_in_library ? 'Remove from Library' : 'Add to Library'}
			>
				<LibraryIcon size={20} />
			</button>

			<button class="icon-action-btn" onclick={onEdit} title="Edit Metadata">
				<Edit3 size={20} />
			</button>
		</div>
	</div>
</div>

<style>
	.hero {
		display: flex;
		gap: 3.5rem;
		margin-bottom: 5rem;
		align-items: stretch;
	}

	.cover-section {
		flex-shrink: 0;
	}

	.cover-wrapper {
		position: relative;
		width: 280px;
		aspect-ratio: 2/3;
		border-radius: 20px;
		overflow: hidden;
		border: 1px solid var(--border-main);
		background: var(--bg-surface);
		cursor: pointer;
		padding: 0;
		box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4);
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
	}

	.cover-wrapper:hover {
		transform: scale(1.02);
		border-color: var(--primary);
	}

	.cover-overlay {
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.2s;
		color: white;
		backdrop-filter: blur(4px);
	}

	.cover-wrapper:hover .cover-overlay {
		opacity: 1;
	}

	.info-section {
		flex: 1;
		min-width: 0;
		display: flex;
		flex-direction: column;
	}

	.top-meta {
		display: flex;
		gap: 0.75rem;
		margin-bottom: 1.25rem;
	}

	.source-tag,
	.status-tag {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		padding: 0.3rem 0.75rem;
		border-radius: 8px;
		background: rgba(255, 255, 255, 0.03);
		border: 1px solid var(--border-main);
		color: var(--text-dim);
	}

	.status-tag {
		background: rgba(var(--primary-rgb), 0.1);
		color: var(--primary);
		border-color: rgba(var(--primary-rgb), 0.2);
	}

	h1 {
		font-size: 3.5rem;
		font-weight: 900;
		margin: 0 0 2rem;
		letter-spacing: -0.04em;
		line-height: 1;
		color: var(--text-main);
	}

	.meta-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		margin-bottom: 3rem;
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		color: var(--text-dim);
		padding: 0.75rem 1.25rem;
		background: rgba(255, 255, 255, 0.02);
		border-radius: 12px;
		border: 1px solid var(--border-main);
	}

	.meta-label {
		display: flex;
		flex-direction: column;
		gap: 0.1rem;
	}

	.meta-label .label {
		font-size: 0.6rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		opacity: 0.5;
	}

	.meta-label .value {
		font-size: 0.9rem;
		font-weight: 600;
		color: var(--text-main);
	}

	.genres {
		color: var(--primary);
	}

	.synopsis-box h3 {
		font-size: 0.875rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		margin: 0 0 1rem;
		color: var(--text-dim);
	}

	.synopsis {
		font-size: 1.0625rem;
		line-height: 1.7;
		color: var(--text-muted);
		margin-bottom: 0.5rem;
		display: -webkit-box;
		-webkit-line-clamp: 5;
		line-clamp: 5;
		-webkit-box-orient: vertical;
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.synopsis.clickable {
		cursor: pointer;
	}

	.synopsis.clickable:hover {
		color: var(--text-main);
	}

	.synopsis.expanded {
		-webkit-line-clamp: unset;
		line-clamp: unset;
		display: block;
		margin-bottom: 1rem;
	}

	.expand-toggle {
		background: none;
		border: none;
		color: var(--primary);
		font-size: 0.8125rem;
		font-weight: 800;
		padding: 0;
		cursor: pointer;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 3.5rem;
		transition: color 0.2s;
	}

	.expand-toggle:hover {
		color: var(--text-main);
		text-decoration: underline;
	}

	.main-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 1rem;
		align-items: center;
		margin-top: auto;
	}

	.primary-btn {
		background: white;
		color: black;
		border: none;
		padding: 0 2.5rem;
		height: 56px;
		border-radius: 16px;
		font-weight: 800;
		font-size: 1rem;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		cursor: pointer;
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		box-shadow: 0 8px 24px rgba(255, 255, 255, 0.1);
		white-space: nowrap;
		flex-shrink: 0;
	}

	.primary-btn:hover {
		transform: translateY(-2px);
		box-shadow: 0 12px 32px rgba(255, 255, 255, 0.15);
		background: #f4f4f5;
	}

	.secondary-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0 1.5rem;
		height: 56px;
		border-radius: 16px;
		font-weight: 700;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
		white-space: nowrap;
		flex-shrink: 0;
	}

	.secondary-btn:hover {
		background: var(--bg-surface-hover);
		border-color: var(--border-bright);
	}

	.icon-action-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
		width: 56px;
		height: 56px;
		border-radius: 16px;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.2s;
	}

	.icon-action-btn:hover {
		border-color: var(--border-bright);
		color: var(--text-main);
	}

	.icon-action-btn.active {
		color: var(--primary);
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}

	@media (max-width: 1100px) {
		h1 {
			font-size: 2.75rem;
		}
		.hero {
			gap: 2.5rem;
		}
		.cover-wrapper {
			width: 220px;
		}
	}

	@media (max-width: 900px) {
		.hero {
			flex-direction: column;
			gap: 2rem;
			align-items: center;
		}
		.info-section {
			width: 100%;
			text-align: center;
		}
		.top-meta,
		.meta-grid,
		.main-actions {
			justify-content: center;
		}
		.meta-grid {
			text-align: left;
		}
		.cover-wrapper {
			width: 240px;
		}
		h1 {
			font-size: 2.25rem;
		}
	}

	@media (max-width: 600px) {
		.primary-btn,
		.secondary-btn {
			width: 100%;
			height: 48px;
		}
		.icon-action-btn {
			flex: 1;
			height: 48px;
		}
	}
</style>
