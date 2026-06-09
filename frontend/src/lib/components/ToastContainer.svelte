<script lang="ts">
	import { toast } from '$lib/services/toast.svelte';
	import CheckCircle2 from 'lucide-svelte/icons/check-circle-2';
	import AlertCircle from 'lucide-svelte/icons/alert-circle';
	import Info from 'lucide-svelte/icons/info';
	import X from 'lucide-svelte/icons/x';
	import { flip } from 'svelte/animate';
	import { fade, fly } from 'svelte/transition';
	import { clsx } from 'clsx';
</script>

<div class="toast-container">
	{#each toast.toasts as t (t.id)}
		<div
			animate:flip={{ duration: 300 }}
			in:fly={{ y: 20, duration: 400 }}
			out:fade={{ duration: 200 }}
			class={clsx('toast', t.type)}
		>
			<div class="icon-box">
				{#if t.type === 'success'}
					<CheckCircle2 size={18} />
				{:else if t.type === 'error'}
					<AlertCircle size={18} />
				{:else}
					<Info size={18} />
				{/if}
			</div>
			<span class="message">{t.message}</span>
			<button class="close-btn" onclick={() => toast.remove(t.id)}>
				<X size={14} />
			</button>
		</div>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		bottom: 2rem;
		right: 2rem;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		pointer-events: none;
	}

	.toast {
		pointer-events: auto;
		min-width: 300px;
		max-width: 450px;
		background: var(--bg-surface);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 1rem;
		display: flex;
		align-items: center;
		gap: 0.75rem;
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
	}

	.icon-box {
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.toast.success .icon-box {
		color: #10b981;
	}
	.toast.error .icon-box {
		color: #ef4444;
	}
	.toast.info .icon-box {
		color: #3b82f6;
	}

	.message {
		flex: 1;
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-main);
		line-height: 1.4;
	}

	.close-btn {
		background: none;
		border: none;
		color: var(--text-dim);
		cursor: pointer;
		padding: 0.25rem;
		border-radius: 4px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all 0.2s;
	}

	.close-btn:hover {
		color: var(--text-main);
		background: rgba(255, 255, 255, 0.05);
	}
</style>
