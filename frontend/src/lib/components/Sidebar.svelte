<script lang="ts">
	import { Settings, Compass, Zap, ChevronRight } from 'lucide-svelte';
	import * as Icons from 'lucide-svelte';
	import { clsx } from 'clsx';

	interface Tab {
		id: string;
		label: string;
		icon: string;
		section_id: string;
		component: string;
	}

	interface Section {
		id: string;
		label: string;
	}

	interface Plugin {
		name: string;
		is_enabled: boolean;
		is_loaded: boolean;
		tabs: Tab[];
		sections: Section[];
	}

	interface Props {
		plugins: Plugin[];
		currentView: string;
		currentPlugin: string;
		currentTabId: string;
		onNavigate: (view: string, plugin?: string, tabId?: string) => void;
	}
	let { plugins, currentView, currentTabId, onNavigate }: Props = $props();

	let allSections = $derived(() => {
		const sections: Section[] = [];
		plugins
			.filter((p) => p.is_loaded)
			.forEach((p) => {
				p.sections.forEach((s) => {
					if (!sections.find((ex) => ex.id === s.id)) {
						sections.push(s);
					}
				});
			});
		return sections;
	});

	function getIcon(name: string) {
		// @ts-expect-error - dynamic icon indexing
		return Icons[name] || Compass;
	}
</script>

<aside class="sidebar">
	<div class="brand">
		<span>Lector</span>
	</div>

	{#each allSections() as section (section.id)}
		<nav class="nav-section">
			<header>{section.label}</header>
			<div class="nav-list">
				{#each plugins.filter((p) => p.is_loaded) as p (p.name)}
					{#each p.tabs.filter((t) => t.section_id === section.id) as t (t.id)}
						{@const Icon = getIcon(t.icon)}
						<button
							class={clsx('nav-item', currentTabId === t.id && 'active')}
							onclick={() => onNavigate(t.component || 'explorer', p.name, t.id)}
						>
							<Icon size={18} />
							<span>{t.label}</span>
							{#if section.id === 'sources' || section.id === 'plugins'}
								<ChevronRight size={14} class="chevron" />
							{/if}
						</button>
					{/each}
				{/each}
			</div>
		</nav>
	{/each}

	<div class="footer">
		<button
			class={clsx('nav-item', currentView === 'settings' && 'active')}
			onclick={() => onNavigate('settings')}
		>
			<Settings size={18} />
			<span>Settings</span>
		</button>
	</div>
</aside>

<style>
	.sidebar {
		width: 260px;
		background: var(--bg-main);
		border-right: 1px solid var(--border-main);
		height: 100vh;
		display: flex;
		flex-direction: column;
		padding: 1.5rem 0.75rem;
	}

	.brand {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0 1rem;
		margin-bottom: 2.5rem;
	}

	.brand span {
		font-size: 1.1rem;
		font-weight: 700;
		letter-spacing: -0.02em;
		color: var(--text-main);
	}

	.nav-section {
		margin-bottom: 2rem;
	}

	.nav-section.plugins {
		flex-grow: 1;
		overflow-y: auto;
	}

	header {
		font-size: 0.7rem;
		font-weight: 600;
		color: var(--text-dim);
		margin-bottom: 0.5rem;
		padding-left: 1rem;
		text-transform: none;
	}

	.nav-list {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}

	.nav-item {
		width: 100%;
		background: none;
		border: none;
		color: var(--text-muted);
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.6rem 1rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
		border-radius: 6px;
		text-align: left;
	}

	.nav-item:hover {
		color: var(--text-main);
		background: var(--bg-surface);
	}

	.nav-item.active {
		color: var(--primary);
		background: var(--bg-surface);
	}

	.chevron {
		margin-left: auto;
		opacity: 0.5;
	}

	.source-item span {
		font-size: 0.875rem;
	}

	.capitalize {
		text-transform: capitalize;
	}

	.footer {
		margin-top: auto;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border-main);
	}
</style>
