export interface SearchItem {
	title: string;
	url: string;
	cover_url: string;
	info: string;
	source?: string;
}

export interface Chapter {
	id: number;
	document_id: number;
	title: string;
	url: string;
	content: string;
	metadata: string;
	order: number;
	status: string;
	is_read: boolean;
}

export interface Document {
	id: number;
	title: string;
	url: string;
	source: string;
	type: 'text' | 'images' | 'stream';
	cover_url: string;
	author: string;
	studio: string;
	synopsis: string;
	genres: string;
	status: string;
	is_in_library: boolean;
	is_archived: boolean;
	is_local: boolean;
	local_path: string;
	group_id: number;
	chapters: Chapter[];
	read_chapters: number;
	total_chapters: number;
}

export interface Group {
	id: number;
	name: string;
	documents: Document[];
}

export interface Plugin {
	id: number;
	name: string;
	is_enabled: boolean;
	tabs?: { id: string; label: string; icon: string; section_id: string; component: string }[];
}

export interface PluginManifest {
	name: string;
	is_enabled: boolean;
	is_loaded: boolean;
	is_verified: boolean;
	tabs: {
		id: string;
		label: string;
		icon: string;
		section_id: string;
		component: string;
	}[];
	sections: {
		id: string;
		label: string;
	}[];
	settings_groups: {
		id: string;
		label: string;
	}[];
	actions: {
		context: string;
		label: string;
		method: string;
		icon: string;
	}[];
	ui_overrides: Record<string, Record<string, string>>;
	permissions: string[];
	capabilities: string[];
	css: string;
}

export interface Bookmark {
	id: number;
	document_id: number;
	chapter_id: number;
	title: string;
	created_at: string;
}

export interface Note {
	id: number;
	document_id: number;
	chapter_id: number;
	content: string;
	quote: string;
	created_at: string;
}

export interface SearchResponse {
	results: SearchItem[];
	errors: string[];
}
