import type { PluginManifest } from './types';

export const baseManifest: PluginManifest = {
	name: 'system',
	is_enabled: true,
	is_loaded: true,
	sections: [{ id: 'workspace', label: 'Workspace' }],
	tabs: [
		{
			id: 'sys:library',
			label: 'Library',
			icon: 'Library',
			section_id: 'workspace',
			component: 'library'
		},
		{
			id: 'sys:history',
			label: 'History',
			icon: 'Clock',
			section_id: 'workspace',
			component: 'history'
		},
		{
			id: 'sys:search',
			label: 'Search',
			icon: 'Search',
			section_id: 'workspace',
			component: 'search'
		},
		{
			id: 'sys:plugins',
			label: 'Plugins',
			icon: 'Zap',
			section_id: 'workspace',
			component: 'plugins'
		}
	],
	settings_groups: [],
	actions: [],
	ui_overrides: {},
	permissions: []
};
