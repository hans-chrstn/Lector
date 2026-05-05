<script lang="ts">
	import type { Snippet } from 'svelte';
	import { clsx } from 'clsx';

	interface Props {
		title: string;
		subtitle?: string;
		titleBadge?: string;
		actions?: Snippet;
		extraHeader?: Snippet;
		children: Snippet;
		containerClass?: string;
	}

	let { title, subtitle, titleBadge, actions, extraHeader, children, containerClass }: Props =
		$props();
</script>

<div class={clsx('base-page', containerClass)}>
	<header class="page-header">
		<div class="header-main">
			<div class="title-area">
				<div class="title-group">
					<h1>{title}</h1>
					{#if titleBadge}
						<span class="title-badge">{titleBadge}</span>
					{/if}
				</div>
				{#if subtitle}
					<p class="subtitle">{subtitle}</p>
				{/if}
			</div>

			{#if actions}
				<div class="header-actions">
					{@render actions()}
				</div>
			{/if}
		</div>

		{#if extraHeader}
			<div class="header-extra">
				{@render extraHeader()}
			</div>
		{/if}
	</header>

	<main class="page-content">
		{@render children()}
	</main>
</div>

<style>
	.base-page {
		padding-bottom: 5rem;
		animation: fadeIn 0.3s ease-out;
		width: 100%;
	}

	.page-header {
		margin-bottom: 2.5rem;
	}

	.header-main {
		display: flex;
		justify-content: space-between;
		align-items: flex-end;
		margin-bottom: 2rem;
		gap: 2rem;
	}

	.title-area {
		flex: 1;
	}

	.title-group {
		display: flex;
		align-items: baseline;
		gap: 1rem;
	}

	h1 {
		font-size: 2.5rem;
		font-weight: 800;
		margin: 0;
		letter-spacing: -0.04em;
		color: var(--text-main);
	}

	.title-badge {
		color: var(--text-dim);
		font-weight: 600;
		font-size: 0.9rem;
	}

	.subtitle {
		color: var(--text-dim);
		font-size: 0.875rem;
		margin: 0.25rem 0 0;
		line-height: 1.5;
	}

	.header-actions {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.header-extra {
		margin-top: 2rem;
	}

	.page-content {
		width: 100%;
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

	@media (max-width: 900px) {
		.header-main {
			flex-direction: column;
			align-items: flex-start;
			gap: 1.5rem;
		}

		.header-actions {
			width: 100%;
		}

		h1 {
			font-size: 2rem;
		}
	}
</style>
