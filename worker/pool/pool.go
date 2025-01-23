package pool

import (
	"context"
	"sync"
	"time"
)

type Worker struct {
	ID          int
	State       string
	Task        Task
	StartTime   time.Time
	ElapsedTime time.Duration
}

type Task func(context.Context)

type WorkerPool struct {
	tasks       chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	poolSize    int
	workers     []*Worker
	lock        sync.Mutex
	activeTasks int
}

// NewWorkerPool creates a new WorkerPool with the specified size.
func NewWorkerPool(poolSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		tasks:    make(chan Task),
		ctx:      ctx,
		cancel:   cancel,
		poolSize: poolSize,
		workers:  make([]*Worker, poolSize),
	}
}

// Start starts the worker pool and creates the worker goroutines.
func (p *WorkerPool) Start() {
	for i := 0; i < p.poolSize; i++ {
		worker := &Worker{
			ID:    i,
			State: "idle",
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
			p.lock.Lock()
			worker.State = "running"
			worker.Task = task
			worker.StartTime = time.Now()
			p.activeTasks++
			p.lock.Unlock()

			ctx, cancel := context.WithTimeout(p.ctx, 2*time.Second) // Set a 2-second timeout
			defer cancel()

			task(ctx)

			p.lock.Lock()
			worker.State = "idle"
			worker.ElapsedTime = time.Since(worker.StartTime)
			p.activeTasks--
			p.lock.Unlock()
		case <-p.ctx.Done():
			return
		}
	}
}

// Submit submits a task to the worker pool.
func (p *WorkerPool) Submit(task Task) {
	p.tasks <- task
}

// Shutdown gracefully shuts down the worker pool, waiting for all tasks to complete.
func (p *WorkerPool) Shutdown() {
	p.cancel()
	close(p.tasks)
	p.wg.Wait()
}

// GetWorkers returns the list of workers.
func (p *WorkerPool) GetWorkers() []*Worker {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.workers
}

// GetWorker returns a worker by ID.
func (p *WorkerPool) GetWorker(id int) *Worker {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.workers[id]
}

// GetTaskCount returns the number of active tasks being processed.
func (p *WorkerPool) GetTaskCount() int {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.activeTasks
}

// GetPoolSize returns the pool size.
func (p *WorkerPool) GetPoolSize() int {
	return p.poolSize
}

// Monitor provides information about the current state of the workers.
func (p *WorkerPool) Monitor() []*Worker {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.workers
}
