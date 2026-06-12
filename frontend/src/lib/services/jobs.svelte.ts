import { api } from './api';

export interface Job {
	id: number;
	type: string;
	status: string;
	progress: number;
	message: string;
	payload: string;
	error: string;
	created_at: string;
	updated_at: string;
}

class JobState {
	activeJobs = $state<Map<number, Job>>(new Map());

	constructor() {
		if (typeof window !== 'undefined') {
			const cleanup = api.connectSSE((job: Job) => {
				if (job.status === 'completed' || job.status === 'failed') {
					this.activeJobs.delete(job.id);
					this.activeJobs = new Map(this.activeJobs);
				} else {
					this.activeJobs.set(job.id, job);
					this.activeJobs = new Map(this.activeJobs);
				}
			});
			window.addEventListener('beforeunload', cleanup);
		}
	}
}

export const jobState = new JobState();
