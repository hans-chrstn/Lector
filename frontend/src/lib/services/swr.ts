type Fetcher<T> = () => Promise<T>;

class SWREngine {
	private cache = new Map<string, unknown>();
	private subscribers = new Map<string, Set<(data: unknown) => void>>();

	async use<T>(key: string, fetcher: Fetcher<T>): Promise<T> {
		const cached = this.cache.get(key) as T | undefined;

		const promise = fetcher().then((data) => {
			this.cache.set(key, data);
			this.notify(key, data);
			return data;
		});

		return cached || (await promise);
	}

	subscribe<T>(key: string, callback: (data: T) => void) {
		if (!this.subscribers.has(key)) {
			this.subscribers.set(key, new Set());
		}
		const wrappedCallback = callback as (data: unknown) => void;
		this.subscribers.get(key)!.add(wrappedCallback);

		const current = this.cache.get(key);
		if (current) callback(current as T);

		return () => this.subscribers.get(key)!.delete(wrappedCallback);
	}

	private notify(key: string, data: unknown) {
		const subs = this.subscribers.get(key);
		if (subs) {
			subs.forEach((cb) => cb(data));
		}
	}
}

export const swr = new SWREngine();
