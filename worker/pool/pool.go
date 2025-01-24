package pool

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type Worker struct {
	ID          int
	State       string
	Task        Task
	StartTime   time.Time
	ElapsedTime time.Duration
	quit        chan struct{} // Channel for signaling worker to quit
}

type Task func(context.Context) error

type WorkerPool struct {
	tasks            chan Task
	wg               sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
	poolSize         int
	workers          []*Worker
	rwLock           sync.RWMutex // Read-write lock
	activeTasks      int
	activeTasksMutex sync.Mutex // Protect activeTasks
	maxTasks         int        // Maximum number of concurrent tasks
}

// NewWorkerPool creates a new WorkerPool with the specified size and maximum concurrent tasks.
func NewWorkerPool(poolSize int, maxTasks int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		tasks:    make(chan Task),
		ctx:      ctx,
		cancel:   cancel,
		poolSize: poolSize,
		workers:  make([]*Worker, poolSize),
		maxTasks: maxTasks,
	}
}

// Start starts the worker pool and creates the worker goroutines.
func (p *WorkerPool) Start() {
	for i := 0; i < p.poolSize; i++ {
		worker := &Worker{
			ID:    i,
			State: "idle",
			quit:  make(chan struct{}),
		}
		p.workers[i] = worker
		p.wg.Add(1)
		go p.worker(worker)
	}
}

// worker is the function executed by each worker goroutine.
func (p *WorkerPool) worker(worker *Worker) {
	defer p.wg.Done()
	for {
		select {
		case task := <-p.tasks:
			p.rwLock.Lock() // Acquire write lock for worker state modifications
			worker.State = "running"
			worker.Task = task
			worker.StartTime = time.Now()
			p.rwLock.Unlock()

			p.activeTasksMutex.Lock()
			p.activeTasks++
			p.activeTasksMutex.Unlock()

			// Execute the task with a timeout
			ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second) // Adjust timeout as needed
			defer cancel()

			err := worker.Task(ctx)
			if err != nil {
				// Handle task errors (e.g., log, retry)
				// ...
				log.Default().Println("Task error:", err)
			}

			p.rwLock.Lock() // Acquire write lock for worker state modifications
			worker.State = "idle"
			worker.ElapsedTime = time.Since(worker.StartTime)
			p.rwLock.Unlock()

			p.activeTasksMutex.Lock()
			p.activeTasks--
			p.activeTasksMutex.Unlock()
		case <-worker.quit: // Check for quit signal
			return
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit submits a task to the worker pool.
func (p *WorkerPool) Submit(task Task) error {
	// Check if the maximum number of concurrent tasks is reached
	p.activeTasksMutex.Lock()
	defer p.activeTasksMutex.Unlock()
	if p.activeTasks >= p.maxTasks {
		return ErrMaxTasksReached // Define a custom error
	}

	p.tasks <- task
	return nil
}

// Shutdown gracefully shuts down the worker pool, waiting for all tasks to complete.
func (p *WorkerPool) Shutdown() {
	p.cancel()
	close(p.tasks)

	// Signal workers to quit
	for _, worker := range p.workers {
		close(worker.quit)
	}

	p.wg.Wait()
}

// GetWorkers returns the list of workers.
func (p *WorkerPool) GetWorkers() []*Worker {
	p.rwLock.RLock() // Acquire read lock
	defer p.rwLock.RUnlock()
	return p.workers
}

// GetWorker returns a worker by ID.
func (p *WorkerPool) GetWorker(id int) *Worker {
	p.rwLock.RLock() // Acquire read lock
	defer p.rwLock.RUnlock()
	return p.workers[id]
}

// GetTaskCount returns the number of active tasks being processed.
func (p *WorkerPool) GetTaskCount() int {
	p.activeTasksMutex.Lock()
	defer p.activeTasksMutex.Unlock()
	return p.activeTasks
}

// GetPoolSize returns the pool size.
func (p *WorkerPool) GetPoolSize() int {
	return p.poolSize
}

// Monitor provides information about the current state of the workers.
func (p *WorkerPool) Monitor() []*Worker {
	p.rwLock.Lock()
	defer p.rwLock.Unlock()
	return p.workers
}

// Error for maximum tasks reached
var ErrMaxTasksReached = errors.New("maximum number of concurrent tasks reached")
