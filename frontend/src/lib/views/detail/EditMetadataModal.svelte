<script lang="ts">
	import Modal from '../../components/Modal.svelte';
	import type { Document } from '$lib/services/api';

	interface Props {
		show: boolean;
		document: Document;
		onClose: () => void;
		onSave: (form: any) => void;
	}

	let { show, document: doc, onClose, onSave }: Props = $props();

	let form = $state({
		title: '',
		author: '',
		synopsis: '',
		genres: '',
		status: ''
	});

	$effect(() => {
		if (show) {
			form = {
				title: doc.title || '',
				author: doc.author || '',
				synopsis: doc.synopsis || '',
				genres: doc.genres || '',
				status: doc.status || ''
			};
		}
	});

	function handleSubmit() {
		onSave(form);
	}
</script>

<Modal {show} title="Edit Metadata" {onClose}>
	<div class="edit-form">
		<div class="field">
			<label for="title">Title</label>
			<input id="title" type="text" bind:value={form.title} />
		</div>
		<div class="field">
			<label for="author">Author</label>
			<input id="author" type="text" bind:value={form.author} />
		</div>
		<div class="field">
			<label for="genres">Genres</label>
			<input
				id="genres"
				type="text"
				bind:value={form.genres}
				placeholder="Action, Adventure, etc."
			/>
		</div>
		<div class="field">
			<label for="status">Status</label>
			<select id="status" bind:value={form.status}>
				<option value="ongoing">Ongoing</option>
				<option value="completed">Completed</option>
				<option value="hiatus">Hiatus</option>
			</select>
		</div>
		<div class="field">
			<label for="synopsis">Synopsis</label>
			<textarea id="synopsis" bind:value={form.synopsis} rows="5"></textarea>
		</div>

		<div class="actions">
			<button class="btn secondary" onclick={onClose}>Cancel</button>
			<button class="btn primary" onclick={handleSubmit}>Save Changes</button>
		</div>
	</div>
</Modal>

<style>
	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 1.25rem;
		padding-top: 0.5rem;
	}
	.field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}
	label {
		font-size: 0.75rem;
		font-weight: 700;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	input,
	textarea,
	select {
		background: var(--bg-main);
		border: 1px solid var(--border-bright);
		color: var(--text-main);
		padding: 0.75rem;
		border-radius: 8px;
		font-size: 0.9375rem;
		outline: none;
	}
	input:focus,
	textarea:focus,
	select:focus {
		border-color: var(--primary);
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		margin-top: 1rem;
	}
	.btn {
		padding: 0.6rem 1.2rem;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.875rem;
		cursor: pointer;
		border: none;
	}
	.btn.primary {
		background: var(--primary);
		color: white;
	}
	.btn.secondary {
		background: var(--bg-surface-hover);
		color: var(--text-main);
	}
</style>
