<script lang="ts">
	import { ChevronLeft, ChevronRight } from 'lucide-svelte';

	interface Props {
		hasPrev: boolean;
		hasNext: boolean;
		currentIdx: number;
		onNavigate: (offset: number) => void;
	}

	let { hasPrev, hasNext, currentIdx, onNavigate }: Props = $props();
</script>

<footer class="chapter-footer">
	<div class="footer-nav">
		<button
			type="button"
			class="nav-btn"
			disabled={!hasPrev}
			onclick={(e) => {
				e.stopPropagation();
				onNavigate(-1);
			}}
		>
			<ChevronLeft size={20} />
			<div class="btn-info">
				<span class="dir">Previous</span>
				<span class="label">Chapter {currentIdx}</span>
			</div>
		</button>
		<button
			type="button"
			class="nav-btn next"
			disabled={!hasNext}
			onclick={(e) => {
				e.stopPropagation();
				onNavigate(1);
			}}
		>
			<div class="btn-info">
				<span class="dir">Next</span>
				<span class="label">Chapter {currentIdx + 2}</span>
			</div>
			<ChevronRight size={20} />
		</button>
	</div>
</footer>

<style>
	.chapter-footer {
		margin-top: 4rem;
		border-top: 1px solid var(--border-main);
	}
	.footer-nav {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		padding-top: 3rem;
	}
	.nav-btn {
		flex: 1;
		display: flex;
		align-items: center;
		gap: 1rem;
		background: var(--surface);
		border: 1px solid var(--border-main);
		padding: 1rem 1.25rem;
		border-radius: 12px;
		color: var(--text);
		transition: all 0.2s ease;
		text-align: left;
	}
	.nav-btn:hover:not(:disabled) {
		border-color: var(--primary);
		color: var(--text-strong);
	}
	.nav-btn:disabled {
		opacity: 0.3;
		cursor: default;
	}
	.nav-btn.next {
		flex-direction: row-reverse;
		text-align: right;
	}
	.btn-info {
		display: flex;
		flex-direction: column;
		gap: 0.15rem;
	}
	.btn-info .dir {
		font-size: 0.65rem;
		font-weight: 800;
		opacity: 0.5;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	.btn-info .label {
		font-size: 0.8125rem;
		font-weight: 600;
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
