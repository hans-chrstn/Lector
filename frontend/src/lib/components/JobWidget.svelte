<script lang="ts">
	import { jobState } from '$lib/services/jobs.svelte';
	import { Loader2 } from 'lucide-svelte';
</script>

{#if jobState.activeJobs.size > 0}
	<div class="job-widget">
		{#each Array.from(jobState.activeJobs.values()) as job (job.id)}
			<div class="job-item">
				<div class="job-header">
					<span class="job-type">{job.type.replace('_', ' ')}</span>
					<Loader2 class="spin" size={14} />
				</div>
				<div class="job-msg">{job.message}</div>
				<div class="progress-bar">
					<div class="progress-fill" style="width: {job.progress}%"></div>
				</div>
			</div>
		{/each}
	</div>
{/if}

<style>
	.job-widget {
		position: fixed;
		bottom: 24px;
		left: 24px;
		width: 300px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		z-index: 9999;
	}
	.job-item {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 12px;
		padding: 12px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}
	.job-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 4px;
	}
	.job-type {
		font-weight: 600;
		font-size: 0.875rem;
		text-transform: capitalize;
		color: var(--text);
	}
	.job-msg {
		font-size: 0.75rem;
		color: var(--muted);
		margin-bottom: 8px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.progress-bar {
		height: 4px;
		background: var(--surface-hover);
		border-radius: 2px;
		overflow: hidden;
	}
	.progress-fill {
		height: 100%;
		background: var(--primary);
		transition: width 0.3s ease;
	}
	:global(.spin) {
		animation: spin 1s linear infinite;
	}
	@keyframes spin {
		100% {
			transform: rotate(360deg);
		}
	}
</style>
