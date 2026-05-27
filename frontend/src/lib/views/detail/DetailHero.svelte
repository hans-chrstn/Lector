<script lang="ts">
	import {
		Play,
		Download,
		Library as LibraryIcon,
		Edit3,
		User,
		BookOpen,
		Tag,
		Zap
	} from 'lucide-svelte';
	import * as Icons from 'lucide-svelte';
	import { clsx } from 'clsx';
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
		onPluginAction
	}: Props = $props();

	let coverInput: HTMLInputElement;
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
				<span>{document.author || 'Unknown Author'}</span>
			</div>
			{#if document.chapters}
				<div class="meta-item">
					<BookOpen size={16} />
					<span>{document.chapters.length} Chapters</span>
				</div>
			{/if}
			<div class="meta-item">
				<Tag size={16} />
				<span class="genres">{document.genres || 'No genres listed'}</span>
			</div>
		</div>

		{#if document.synopsis}
			<p class="synopsis">{document.synopsis}</p>
		{/if}

		<div class="main-actions">
			{#if isReady}
				<button class="primary-btn" onclick={onReadAction}>
					<Play size={18} fill="currentColor" />
					<span>{hasStarted ? 'Continue reading' : 'Start reading'}</span>
				</button>

				{#if document.is_in_library}
					<!-- Core Export Button with dynamic check -->
					{#if !availableActions.some((a) => a.label === 'Export EPUB' || a.label === 'Downloaded')}
						<a
							href={`${window.location.origin}/api/documents/${document.id}/export?format=epub`}
							class="secondary-btn"
							download={`${document.title}.epub`}
							rel="external"
						>
							<Download size={18} />
							<span>Export EPUB</span>
						</a>
					{/if}
				{/if}

				{#each availableActions as action, i (i)}
					{@const Icon = (Icons as any)[action.icon] || Zap}
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
		gap: 3rem;
		margin-bottom: 4rem;
		align-items: flex-start;
	}

	.cover-section {
		flex-shrink: 0;
	}

	.cover-wrapper {
		position: relative;
		width: 240px;
		aspect-ratio: 2/3;
		border-radius: 16px;
		overflow: hidden;
		border: 1px solid var(--border-main);
		background: var(--bg-surface);
		cursor: pointer;
		padding: 0;
	}

	.cover-overlay {
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0;
		transition: opacity 0.2s;
		color: white;
	}

	.cover-wrapper:hover .cover-overlay {
		opacity: 1;
	}

	.info-section {
		flex: 1;
		min-width: 0;
	}

	.top-meta {
		display: flex;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.source-tag,
	.status-tag {
		font-size: 0.7rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		padding: 0.25rem 0.6rem;
		border-radius: 6px;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-dim);
	}

	.status-tag {
		background: rgba(var(--primary-rgb), 0.1);
		color: var(--primary);
		border-color: rgba(var(--primary-rgb), 0.2);
	}

	h1 {
		font-size: 2.5rem;
		font-weight: 800;
		margin: 0 0 1.5rem;
		letter-spacing: -0.03em;
		line-height: 1.1;
	}

	.meta-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		color: var(--text-dim);
		font-size: 0.9rem;
		font-weight: 500;
	}

	.genres {
		color: var(--text-main);
	}

	.synopsis {
		font-size: 1rem;
		line-height: 1.6;
		color: var(--text-muted);
		margin-bottom: 2.5rem;
		display: -webkit-box;
		-webkit-line-clamp: 4;
		line-clamp: 4;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.main-actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.primary-btn {
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0 1.5rem;
		height: 48px;
		border-radius: 12px;
		font-weight: 700;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.secondary-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0 1.25rem;
		height: 48px;
		border-radius: 12px;
		font-weight: 700;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
	}

	.icon-action-btn {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		color: var(--text-muted);
		width: 48px;
		height: 48px;
		border-radius: 12px;
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

	@media (max-width: 900px) {
		.hero {
			flex-direction: column;
			gap: 2rem;
		}
		.cover-wrapper {
			width: 180px;
		}
		h1 {
			font-size: 2rem;
		}
	}
</style>
