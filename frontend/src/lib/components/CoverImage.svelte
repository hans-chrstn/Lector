<script lang="ts">
	import { onMount } from 'svelte';
	import { ImageOff, Camera } from 'lucide-svelte';
	import { api } from '$lib/services/api';
	import { clsx } from 'clsx';

	interface Props {
		src: string;
		alt: string;
		className?: string;
		isHero?: boolean;
	}

	let { src, alt, className, isHero = false }: Props = $props();

	let error = $state(false);
	let loading = $state(true);
	let previousSrc = $state('');
	let imgEl = $state<HTMLImageElement>();

	onMount(() => {
		if (imgEl?.complete && imgEl.naturalWidth > 0) {
			loading = false;
		}
	});

	$effect(() => {
		if (src !== previousSrc) {
			previousSrc = src;
			error = false;
			loading = true;
		}
	});

	function handleLoad() {
		if (imgEl && imgEl.naturalWidth > 0) {
			loading = false;
		} else if (imgEl) {
			error = true;
			loading = false;
		}
	}
</script>

<div class={clsx('cover-container', className, isHero && 'hero-size')}>
	{#if src && !error}
		<img
			bind:this={imgEl}
			src={api.getProxyImage(src)}
			{alt}
			class={clsx('cover-img', !loading && 'loaded')}
			onload={handleLoad}
			onerror={() => {
				error = true;
				loading = false;
			}}
		/>
	{/if}

	{#if error || !src}
		<div class="placeholder">
			{#if isHero}
				<Camera size={40} strokeWidth={1.5} />
				<span>No Cover Available</span>
			{:else}
				<ImageOff size={24} strokeWidth={1.5} />
			{/if}
		</div>
	{/if}

	{#if loading && src && !error}
		<div class="shimmer"></div>
	{/if}
</div>

<style>
	.cover-container {
		position: relative;
		width: 100%;
		height: 100%;
		background: var(--bg-surface-hover);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.cover-img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		object-position: center top;
		opacity: 0;
		transition: opacity 0.3s ease;
	}

	.cover-img.loaded {
		opacity: 1;
	}

	.placeholder {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		color: var(--text-dim);
		text-align: center;
		padding: 1rem;
	}

	.placeholder span {
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.shimmer {
		position: absolute;
		inset: 0;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.05), transparent);
		animation: shimmer 1.5s infinite;
	}

	@keyframes shimmer {
		0% {
			transform: translateX(-100%);
		}
		100% {
			transform: translateX(100%);
		}
	}
</style>
