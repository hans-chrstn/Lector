export interface Toast {
	id: string;
	message: string;
	type: 'success' | 'error' | 'info';
	duration: number;
}

class ToastService {
	toasts = $state<Toast[]>([]);

	add(message: string, type: 'success' | 'error' | 'info' = 'info', duration = 3000) {
		const id = Math.random().toString(36).substring(2);
		this.toasts.push({ id, message, type, duration });

		if (duration > 0) {
			setTimeout(() => {
				this.remove(id);
			}, duration);
		}
	}

	remove(id: string) {
		this.toasts = this.toasts.filter((t) => t.id !== id);
	}

	success(message: string, duration = 3000) {
		this.add(message, 'success', duration);
	}

	error(message: string, duration = 3000) {
		this.add(message, 'error', duration);
	}

	info(message: string, duration = 3000) {
		this.add(message, 'info', duration);
	}
}

export const toast = new ToastService();
