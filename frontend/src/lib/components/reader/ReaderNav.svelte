<script lang="ts">
	import {
		X,
		Menu,
		Library,
		BookmarkPlus,
		Layout,
		BookText,
		Volume2,
		VolumeX
	} from 'lucide-svelte';
	import { clsx } from 'clsx';

	interface Props {
		showUI: boolean;
		docTitle: string;
		isSpeaking: boolean;
		onToggleSidebar: () => void;
		onToggleBookmarks: () => void;
		onToggleSettings: () => void;
		onToggleNotes: () => void;
		onToggleTTS: () => void;
		onClose: () => void;
	}

	let {
		showUI,
		docTitle,
		isSpeaking,
		onToggleSidebar,
		onToggleBookmarks,
		onToggleSettings,
		onToggleNotes,
		onToggleTTS,
		onClose
	}: Props = $props();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
<nav class={clsx('reader-nav', showUI && 'visible')} onclick={(e) => e.stopPropagation()}>
	<div class="nav-left">
		<button type="button" class="icon-btn" onclick={onToggleSidebar} title="Contents">
			<Menu size={18} />
		</button>
		<button type="button" class="icon-btn" onclick={onClose} title="Library">
			<Library size={18} />
		</button>
		<button type="button" class="icon-btn" onclick={onToggleBookmarks} title="Bookmarks">
			<BookmarkPlus size={18} />
		</button>
		<button type="button" class="icon-btn" onclick={onToggleTTS} title="Read Aloud">
			{#if isSpeaking}
				<VolumeX size={18} color="var(--primary)" />
			{:else}
				<Volume2 size={18} />
			{/if}
		</button>
	</div>
	<div class="nav-center">
		<span class="document-title">{docTitle}</span>
	</div>
	<div class="nav-right">
		<button type="button" class="icon-btn" onclick={onToggleSettings} title="Settings">
			<Layout size={18} />
		</button>
		<button type="button" class="icon-btn" onclick={onToggleNotes} title="Notebook">
			<BookText size={18} />
		</button>
		<button type="button" class="icon-btn" onclick={onClose} title="Exit">
			<X size={18} />
		</button>
	</div>
</nav>

<style>
	.reader-nav {
		height: 56px;
		background: var(--surface);
		backdrop-filter: blur(16px);
		border-bottom: 1px solid var(--border-main);
		display: flex;
		align-items: center;
		padding: 0 1rem;
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 160;
		opacity: 0;
		transform: translateY(-10px);
		transition: all 0.2s ease;
		pointer-events: none;
	}
	.reader-nav.visible {
		opacity: 1;
		transform: translateY(0);
		pointer-events: auto;
	}
	.nav-left,
	.nav-right {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		flex: 1;
	}
	.nav-right {
		justify-content: flex-end;
	}
	.nav-center {
		flex: 2;
		text-align: center;
		overflow: hidden;
	}
	.document-title {
		display: block;
		font-weight: 600;
		font-size: 0.8125rem;
		color: var(--text-strong);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.icon-btn {
		padding: 0.5rem;
		border-radius: 6px;
		display: flex;
		transition: background 0.2s;
		color: var(--text);
	}
	.icon-btn:hover {
		background: rgba(128, 128, 128, 0.1);
		color: var(--text-strong);
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
</style>
