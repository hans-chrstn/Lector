<script lang="ts">
	interface Props {
		value: number;
		total: number;
		size?: number;
		stroke?: number;
		showText?: boolean;
	}
	let { value, total, size = 24, stroke = 2, showText = false }: Props = $props();

	let percent = $derived(total > 0 ? Math.min(100, Math.max(0, (value / total) * 100)) : 0);
	let radius = $derived((size - stroke) / 2);
	let circumference = $derived(radius * 2 * Math.PI);
	let offset = $derived(circumference - (percent / 100) * circumference);
</script>

<div class="progress-ring" style="width: {size}px; height: {size}px;">
	<svg width={size} height={size}>
		<circle
			class="bg"
			stroke="currentColor"
			stroke-width={stroke}
			fill="transparent"
			r={radius}
			cx={size / 2}
			cy={size / 2}
		/>
		<circle
			class="fg"
			class:complete={percent >= 100}
			stroke="var(--primary)"
			stroke-width={stroke}
			stroke-dasharray="{circumference} {circumference}"
			style="stroke-dashoffset: {offset}"
			stroke-linecap="round"
			fill="transparent"
			r={radius}
			cx={size / 2}
			cy={size / 2}
		/>
	</svg>
	{#if showText}
		<span class="percentage">{Math.round(percent)}%</span>
	{/if}
</div>

<style>
	.progress-ring {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		position: relative;
	}

	svg {
		transform: rotate(-90deg);
	}

	circle {
		transition: stroke-dashoffset 0.35s;
	}

	circle.bg {
		opacity: 0.1;
	}

	circle.fg.complete {
		stroke: #10b981;
	}

	.percentage {
		position: absolute;
		font-size: 0.65rem;
		font-weight: 800;
		color: var(--text-main);
	}
</style>
