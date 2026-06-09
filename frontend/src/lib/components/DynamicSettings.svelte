<script lang="ts">
	import Loader2 from 'lucide-svelte/icons/loader-2';
	import AlertCircle from 'lucide-svelte/icons/alert-circle';
	import { toast } from '$lib/services/toast.svelte';

	interface Props {
		pluginName: string;
		groupId: string;
	}
	let { pluginName, groupId }: Props = $props();

	let schema = $state<any>(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let formData = $state<Record<string, string>>({});

	$effect(() => {
		if (pluginName && groupId) {
			fetchSchema();
		}
	});

	const getBase = () => {
		if (typeof window !== 'undefined') return window.location.origin;
		return 'http://localhost:3000';
	};

	async function fetchSchema() {
		loading = true;
		error = null;
		try {
			const res = await fetch(`${getBase()}/api/plugins/${pluginName}/rpc/get_settings_schema`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(groupId)
			});
			if (!res.ok) throw new Error(await res.text());
			const data = await res.json();
			if (data && data.components) {
				const initialData: Record<string, string> = {};
				for (const comp of data.components) {
					if (comp.type === 'TextInput') {
						initialData[comp.id] = comp.props?.defaultValue || '';
					}
				}
				formData = initialData;
			}
			schema = data;
		} catch (err: any) {
			error = err.message || 'Failed to load settings interface';
		} finally {
			loading = false;
		}
	}

	async function handleAction(method: string) {
		try {
			const res = await fetch(`${getBase()}/api/plugins/${pluginName}/rpc/${method}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(formData)
			});

			if (!res.ok) {
				const text = await res.text();
				toast.error(text || `Action failed (${res.status})`);
				return;
			}

			const data = await res.json();
			if (data.message) toast.success(data.message);
		} catch (e: any) {
			console.error('Settings action error:', e);
			toast.error('Failed to save settings');
		}
	}
</script>

<div class="dynamic-settings">
	{#if loading && !schema}
		<div class="loader-box">
			<Loader2 size={24} class="spin" />
		</div>
	{:else if error}
		<div class="error-box">
			<AlertCircle size={20} />
			<p>{error}</p>
		</div>
	{:else if schema && schema.components}
		{#if schema.title}
			<h4>{schema.title}</h4>
		{/if}

		<div class="settings-form">
			{#each schema.components as component (component.id)}
				{#if component.type === 'TextInput'}
					<div class="input-field">
						{#if component.props?.label}
							<label for={component.id}>{component.props.label}</label>
						{/if}
						<input
							id={component.id}
							type={component.props?.type || 'text'}
							placeholder={component.props?.placeholder || ''}
							bind:value={formData[component.id]}
						/>
					</div>
				{:else if component.type === 'Button'}
					<button class="save-btn" onclick={() => handleAction(component.props.method)}>
						{component.props?.label || 'Submit'}
					</button>
				{/if}
			{/each}
		</div>
	{/if}
</div>

<style>
	.dynamic-settings {
		background: rgba(255, 255, 255, 0.02);
		border: 1px solid var(--border-main);
		border-radius: 12px;
		padding: 1.5rem;
		margin-top: 1rem;
	}

	.dynamic-settings h4 {
		margin: 0 0 1.5rem 0;
		color: var(--text-main);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.loader-box,
	.error-box {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		color: var(--text-dim);
	}

	.error-box {
		color: #ef4444;
	}

	.settings-form {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.input-field {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.input-field label {
		font-size: 0.85rem;
		font-weight: 600;
		color: var(--text-dim);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.input-field input {
		background: rgba(0, 0, 0, 0.2);
		border: 1px solid var(--border-main);
		padding: 0.75rem 1rem;
		border-radius: 8px;
		color: var(--text-main);
		font-size: 0.95rem;
		outline: none;
		transition: border-color 0.2s;
	}

	.input-field input:focus {
		border-color: var(--text-main);
	}

	.save-btn {
		margin-top: 0.5rem;
		background: var(--text-main);
		color: var(--bg-main);
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		font-weight: 700;
		cursor: pointer;
		align-self: flex-start;
		transition: opacity 0.2s;
	}

	.save-btn:hover {
		opacity: 0.9;
	}
</style>
