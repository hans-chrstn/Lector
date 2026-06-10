import type {
	SearchItem,
	SearchResponse,
	Chapter,
	Document as LectorDocument,
	Group,
	Plugin,
	PluginManifest,
	Bookmark,
	Note
} from './types';

export * from './types';

const getBase = () => {
	if (typeof window !== 'undefined') return window.location.origin;
	return 'http://localhost:3000';
};

export interface ReadingStat {
	id: number;
	date: string;
	read_seconds: number;
	documents_read: number;
	chapters_read: number;
}

export const api = {
	async getPlugins(): Promise<string[]> {
		return fetch(`${getBase()}/api/plugins`).then((r) => r.json());
	},
	async getAllPlugins(): Promise<Plugin[]> {
		return fetch(`${getBase()}/api/plugins/all`).then((r) => r.json());
	},
	async getPluginsManifest(): Promise<PluginManifest[]> {
		return fetch(`${getBase()}/api/plugins/manifest`).then((r) => r.json());
	},
	async togglePlugin(name: string): Promise<Plugin> {
		return fetch(`${getBase()}/api/plugins/${name}/toggle`, { method: 'POST' }).then((r) =>
			r.json()
		);
	},
	async uploadPlugin(file: File, name?: string): Promise<{ name: string }> {
		const formData = new FormData();
		formData.append('plugin', file);
		if (name) formData.append('name', name);
		return fetch(`${getBase()}/api/plugins/upload`, {
			method: 'POST',
			body: formData
		}).then((r) => r.json());
	},
	async reorderPlugins(names: string[]): Promise<void> {
		await fetch(`${getBase()}/api/plugins/reorder`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(names)
		});
	},
	async deletePlugin(name: string): Promise<void> {
		await fetch(`${getBase()}/api/plugins/${name}`, { method: 'DELETE' });
	},
	async getActivePlugins(): Promise<string[]> {
		return fetch(`${getBase()}/api/plugins`).then((r) => r.json());
	},
	async getGroups(): Promise<Group[]> {
		return fetch(`${getBase()}/api/groups`).then((r) => r.json());
	},
	async createGroup(name: string): Promise<Group> {
		return fetch(`${getBase()}/api/groups`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name })
		}).then((r) => r.json());
	},
	async deleteGroup(id: number): Promise<void> {
		await fetch(`${getBase()}/api/groups/${id}`, { method: 'DELETE' });
	},
	async getDocuments(archived = false): Promise<LectorDocument[]> {
		return fetch(`${getBase()}/api/documents?archived=${archived}`).then((r) => r.json());
	},
	async searchLibrary(query: string): Promise<LectorDocument[]> {
		return fetch(`${getBase()}/api/documents/search?q=${encodeURIComponent(query)}`).then((r) =>
			r.json()
		);
	},
	async getDocument(id: number): Promise<LectorDocument> {
		return fetch(`${getBase()}/api/documents/${id}`).then((r) => r.json());
	},
	async ensureDocument(url: string, source: string, force?: boolean): Promise<LectorDocument> {
		const endpoint = force
			? `${getBase()}/api/documents/ensure?force=true`
			: `${getBase()}/api/documents/ensure`;
		return fetch(endpoint, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ url, source })
		}).then(async (r) => {
			if (!r.ok) {
				const err = await r.json().catch(() => ({ error: r.statusText }));
				throw new Error(err.error || 'Failed to load document');
			}
			return r.json();
		});
	},
	async refreshDocument(id: number): Promise<LectorDocument> {
		return fetch(`${getBase()}/api/documents/${id}/refresh`, {
			method: 'POST'
		}).then((r) => r.json());
	},
	async batchRefreshDocuments(ids: number[]): Promise<void> {
		await fetch(`${getBase()}/api/documents/batch/refresh`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids })
		});
	},
	async uploadBook(file: File): Promise<LectorDocument> {
		const formData = new FormData();
		formData.append('book', file);
		return fetch(`${getBase()}/api/upload`, {
			method: 'POST',
			body: formData
		}).then((r) => r.json());
	},
	async migrateDocument(id: number, targetUrl: string, source: string): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}/migrate`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ url: targetUrl, source })
		});
	},
	async updateMetadata(id: number, data: Partial<LectorDocument>): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}/metadata`, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
	},
	async updateCover(id: number, file: File): Promise<{ url: string }> {
		const formData = new FormData();
		formData.append('cover', file);
		return fetch(`${getBase()}/api/documents/${id}/cover`, {
			method: 'POST',
			body: formData
		}).then((r) => r.json());
	},
	async toggleLibrary(id: number, inLibrary: boolean): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}/library?is_in_library=${inLibrary}`, {
			method: 'POST'
		});
	},
	async archiveDocument(id: number): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}/archive`, { method: 'POST' });
	},
	async moveDocument(id: number, groupId: number): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}/move`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ group_id: groupId })
		});
	},
	async deleteDocument(id: number): Promise<void> {
		await fetch(`${getBase()}/api/documents/${id}`, { method: 'DELETE' });
	},
	async batchDeleteDocuments(ids: number[]): Promise<void> {
		await fetch(`${getBase()}/api/documents/batch`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids })
		});
	},
	async batchMoveDocuments(ids: number[], groupId: number): Promise<void> {
		await fetch(`${getBase()}/api/documents/batch/move`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids, group_id: groupId })
		});
	},
	async batchArchiveDocuments(ids: number[], archive: boolean): Promise<void> {
		await fetch(`${getBase()}/api/documents/batch/archive`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids, archive })
		});
	},
	async batchMarkReadDocuments(ids: number[], isRead: boolean): Promise<void> {
		await fetch(`${getBase()}/api/documents/batch/mark-read`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids, is_read: isRead })
		});
	},
	async batchUpdateChapters(ids: number[], isRead: boolean): Promise<void> {
		await fetch(`${getBase()}/api/chapters/batch`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids, is_read: isRead })
		});
	},
	async getChapter(id: number): Promise<Chapter> {
		return fetch(`${getBase()}/api/chapters/${id}`).then((r) => r.json());
	},
	async syncProgress(data: {
		document_id: number;
		chapter_id: number;
		scroll_pos: number;
		client_updated_at: number;
	}): Promise<void> {
		await fetch(`${getBase()}/api/progress`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(data)
		});
	},
	async getDocumentProgress(
		id: number
	): Promise<{ chapter_id: number; scroll_pos: number } | null> {
		return fetch(`${getBase()}/api/documents/${id}/progress`).then((r) => {
			if (r.status === 404) return null;
			return r.json();
		});
	},
	async search(source: string, query: string): Promise<SearchResponse> {
		return fetch(
			`${getBase()}/api/search?q=${encodeURIComponent(query)}&plugin=${encodeURIComponent(source)}`
		).then((r) => r.json());
	},
	async getDocumentPopular(plugin: string, page = 1): Promise<SearchItem[]> {
		return fetch(`${getBase()}/api/plugins/${plugin}/popular?page=${page}`).then((r) => r.json());
	},
	async getDocumentDirectory(plugin: string, id: string, page = 1): Promise<SearchItem[]> {
		return fetch(`${getBase()}/api/plugins/${plugin}/directory/${encodeURIComponent(id)}?page=${page}`).then((r) => r.json());
	},
	async getDocumentLatest(plugin: string, page = 1): Promise<SearchItem[]> {
		return fetch(`${getBase()}/api/plugins/${plugin}/latest?page=${page}`).then((r) => r.json());
	},
	async getHistory(): Promise<LectorDocument[]> {
		return fetch(`${getBase()}/api/history`).then((r) => r.json());
	},
	async deleteHistory(id: number): Promise<void> {
		await fetch(`${getBase()}/api/history/${id}`, { method: 'DELETE' });
	},
	async clearHistory(): Promise<void> {
		await fetch(`${getBase()}/api/history`, { method: 'DELETE' });
	},
	async batchDeleteHistory(ids: number[]): Promise<void> {
		await fetch(`${getBase()}/api/history/batch`, {
			method: 'DELETE',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ ids })
		});
	},
	async getBookmarks(documentId: number): Promise<Bookmark[]> {
		return fetch(`${getBase()}/api/documents/${documentId}/bookmarks`).then((r) => r.json());
	},
	async getLibraryPaths(): Promise<
		{ id: number; path: string; pattern: string; is_system: boolean }[]
	> {
		return fetch(`${getBase()}/api/library/paths`).then((r) => r.json());
	},
	async addLibraryPath(path: string, pattern: string): Promise<void> {
		await fetch(`${getBase()}/api/library/paths`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ path, pattern })
		});
	},
	async deleteLibraryPath(id: number): Promise<void> {
		await fetch(`${getBase()}/api/library/paths/${id}`, { method: 'DELETE' });
	},
	async scanLibrary(): Promise<void> {
		await fetch(`${getBase()}/api/library/scan`, { method: 'POST' });
	},
	async addBookmark(documentId: number, chapterId: number, title: string): Promise<void> {
		await fetch(`${getBase()}/api/bookmarks`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ document_id: documentId, chapter_id: chapterId, title })
		});
	},
	async deleteBookmark(id: number): Promise<void> {
		await fetch(`${getBase()}/api/bookmarks/${id}`, { method: 'DELETE' });
	},
	async getNotes(documentId: number): Promise<Note[]> {
		return fetch(`${getBase()}/api/documents/${documentId}/notes`).then((r) => r.json());
	},
	async addNote(
		documentId: number,
		chapterId: number,
		content: string,
		quote: string
	): Promise<void> {
		await fetch(`${getBase()}/api/notes`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ document_id: documentId, chapter_id: chapterId, content, quote })
		});
	},
	async deleteNote(id: number): Promise<void> {
		await fetch(`${getBase()}/api/notes/${id}`, { method: 'DELETE' });
	},
	getProxyImage(url: string): string {
		if (!url || url.startsWith('blob:') || url.startsWith('/')) return url;
		return `/api/proxy-image?url=${encodeURIComponent(url)}`;
	},
	async trackAnalytics(type: string, value: number) {
		await fetch(`${getBase()}/api/analytics/track`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ type, value })
		});
	},
	async getAnalytics(): Promise<ReadingStat[]> {
		return fetch(`${getBase()}/api/analytics`).then((r) => r.json());
	}
};
