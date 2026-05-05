<script lang="ts">
	import type { Document, SearchItem, PluginManifest } from '$lib/services/api';
	import EditMetadataModal from './EditMetadataModal.svelte';
	import MigrateDocumentModal from './MigrateDocumentModal.svelte';

	interface Props {
		document: Document;
		sources: string[];
		pluginManifests: PluginManifest[];
		showEdit: boolean;
		onCloseEdit: () => void;
		onSaveEdit: (form: any) => void;
		showMigrate: boolean;
		onCloseMigrate: () => void;
		onSelectMigrate: (result: SearchItem, source: string) => void;
	}

	let {
		document,
		sources,
		showEdit,
		onCloseEdit,
		onSaveEdit,
		showMigrate,
		onCloseMigrate,
		onSelectMigrate
	}: Props = $props();
</script>

<EditMetadataModal show={showEdit} {document} onClose={onCloseEdit} onSave={onSaveEdit} />

<MigrateDocumentModal
	show={showMigrate}
	{sources}
	onClose={onCloseMigrate}
	onSelect={onSelectMigrate}
/>
