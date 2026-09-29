package worker

import "sync"

type projectOperationLock struct {
	mutex sync.Mutex
	refs  int
}

func (w *SandboxWorker) lockProject(projectID string) func() {
	w.projectLocksMu.Lock()
	lock := w.projectLocks[projectID]
	if lock == nil {
		lock = &projectOperationLock{}
		w.projectLocks[projectID] = lock
	}
	lock.refs++
	w.projectLocksMu.Unlock()

	lock.mutex.Lock()
	return func() {
		lock.mutex.Unlock()
		w.projectLocksMu.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(w.projectLocks, projectID)
		}
		w.projectLocksMu.Unlock()
	}
}
