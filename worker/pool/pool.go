package pool

import (
	"context"
	"sync"
)

type Task func(context.Context)

type WorkerPool struct {
	tasks    chan Task
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	poolSize int
}

func NewWorkerPool(poolSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		tasks:    make(chan Task),
		ctx:      ctx,
		cancel:   cancel,
		poolSize: poolSize,
	}
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}
			task(wp.ctx)
		case <-wp.ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.poolSize; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
	close(wp.tasks)
}

func (wp *WorkerPool) Submit(task Task) {
	wp.tasks <- task
}
