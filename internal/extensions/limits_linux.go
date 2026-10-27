package extensions

import "syscall"

func limitWorker() error {
	for resource, limit := range map[int]uint64{syscall.RLIMIT_DATA: 256 << 20, syscall.RLIMIT_CPU: 7, syscall.RLIMIT_NOFILE: 64, syscall.RLIMIT_CORE: 0} {
		if err := syscall.Setrlimit(resource, &syscall.Rlimit{Cur: limit, Max: limit}); err != nil {
			return err
		}
	}
	return nil
}
