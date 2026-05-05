<script lang="ts">
	interface Settings {
		fontSize: number;
		fontFamily: string;
		theme: 'slate' | 'sepia' | 'light';
		lineHeight: number;
		paragraphSpacing: number;
		sideMargin: number;
		topMargin: number;
		readingMode: 'paged' | 'infinite';
	}

	interface Props {
		settings: Settings;
	}

	let { settings = $bindable() }: Props = $props();
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="settings-panel" onclick={(e) => e.stopPropagation()}>
	<header>Settings</header>
	<div class="setting-group">
		<div class="toggle-row">
			{#each ['slate', 'sepia', 'light'] as t (t)}
				<button
					type="button"
					class:active={settings.theme === t}
					onclick={() => (settings.theme = t as any)}>{t}</button
				>
			{/each}
		</div>
	</div>
	<div class="setting-group">
		<span class="label">Typeface</span>
		<select bind:value={settings.fontFamily}>
			<option value="serif">Iowan Old Style</option>
			<option value="sans">Geist Sans</option>
			<option value="mono">Geist Mono</option>
			<option value="system">System</option>
		</select>
	</div>
	<div class="setting-row">
		<div class="sub-setting">
			<span class="label">Size</span><input
				type="number"
				bind:value={settings.fontSize}
				min="12"
				max="48"
			/>
		</div>
		<div class="sub-setting">
			<span class="label">Spacing</span><input
				type="number"
				step="0.1"
				bind:value={settings.paragraphSpacing}
				min="1"
				max="4"
			/>
		</div>
	</div>
	<div class="setting-group border-t">
		<span class="label">Mode</span>
		<div class="toggle-row">
			<button
				type="button"
				class:active={settings.readingMode === 'paged'}
				onclick={() => (settings.readingMode = 'paged')}>Paged</button
			>
			<button
				type="button"
				class:active={settings.readingMode === 'infinite'}
				onclick={() => (settings.readingMode = 'infinite')}>Infinite</button
			>
		</div>
	</div>
</div>

<style>
	.settings-panel {
		position: fixed;
		top: 64px;
		right: 1rem;
		background: var(--surface);
		border: 1px solid var(--border-main);
		padding: 1.25rem;
		border-radius: 12px;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
		z-index: 170;
		width: 280px;
		display: flex;
		flex-direction: column;
		gap: 1rem;
		color: var(--text);
	}
	.settings-panel header {
		font-size: 0.65rem;
		font-weight: 800;
		text-transform: uppercase;
		letter-spacing: 1px;
	}
	.setting-group {
		display: flex;
		flex-direction: column;
		gap: 0.4rem;
	}
	.setting-row {
		display: flex;
		gap: 0.75rem;
	}
	.sub-setting {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}
	.label {
		font-size: 0.65rem;
		font-weight: 600;
		opacity: 0.7;
	}
	.toggle-row {
		display: flex;
		background: rgba(0, 0, 0, 0.08);
		padding: 2px;
		border-radius: 6px;
	}
	.toggle-row button {
		flex: 1;
		padding: 0.35rem;
		font-size: 0.7rem;
		border-radius: 4px;
		color: var(--text);
		text-align: center;
	}
	.toggle-row button.active {
		background: var(--bg);
		color: var(--text-strong);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}
	.border-t {
		border-top: 1px solid var(--border-main);
		padding-top: 0.5rem;
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
	select,
	input {
		background: var(--bg);
		color: var(--text-strong);
		border: 1px solid var(--border-main);
		border-radius: 4px;
		padding: 4px;
		font-size: 0.875rem;
	}
</style>
