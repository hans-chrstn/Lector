<script lang="ts">
	import '@fontsource/geist-sans/index.css';
	import '@fontsource/geist-mono/index.css';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	let { children } = $props();
</script>

{@render children()}
<ToastContainer />
