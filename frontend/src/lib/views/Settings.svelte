<script lang="ts">
	import { ShieldAlert } from 'lucide-svelte';
	import BasePage from '../components/base/BasePage.svelte';

	interface Props {
		sources: string[];
		onRefreshSources: () => void;
	}
	let {}: Props = $props();
</script>

<BasePage title="Settings" subtitle="Configure your instance preferences">
	<div class="settings-grid">
		<!-- System Section -->
		<section class="settings-card">
			<header>
				<div class="header-left">
					<ShieldAlert size={20} class="icon" />
					<h2>System & Safety</h2>
				</div>
			</header>
			<div class="system-opts">
				<div class="opt-item">
					<div class="opt-info">
						<span class="opt-label">Clear Cache</span>
						<span class="opt-desc">Delete all cached discovery results and images</span>
					</div>
					<button class="secondary-btn">Clear Data</button>
				</div>
				<div class="opt-item">
					<div class="opt-info">
						<span class="opt-label">Version</span>
						<span class="opt-desc">Production Build v0.3.4</span>
					</div>
				</div>
			</div>
		</section>
	</div>
</BasePage>

<style>
	.settings-grid {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.settings-card {
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		overflow: hidden;
	}

	.settings-card header {
		padding: 1.25rem 1.5rem;
		border-bottom: 1px solid var(--border-main);
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: rgba(255, 255, 255, 0.02);
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}
	h2 {
		font-size: 0.9375rem;
		font-weight: 600;
		color: var(--text-main);
		margin: 0;
	}

	.system-opts {
		padding: 1rem 1.5rem;
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}
	.opt-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	.opt-info {
		display: flex;
		flex-direction: column;
		gap: 0.125rem;
	}
	.opt-label {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-main);
	}
	.opt-desc {
		font-size: 0.75rem;
		color: var(--text-dim);
	}

	.secondary-btn {
		background: var(--bg-main);
		border: 1px solid var(--border-main);
		color: var(--text-main);
		padding: 0.4rem 0.875rem;
		border-radius: 6px;
		font-size: 0.8125rem;
		font-weight: 600;
		cursor: pointer;
	}
</style>
