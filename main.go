package main

import (
	"fmt"
	"sync"
	"time"
)

type Worker struct {
	id         int
	stopSignal chan bool
}

func (w *Worker) Start(inputChannel <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case data := <-inputChannel:
			fmt.Printf("Worker %d: received data - %s\n", w.id, data)
		case <-w.stopSignal:
			fmt.Printf("Worker %d: stopping\n", w.id)
			return
		}
	}
}

type WorkerPool struct {
	inputChannel chan string
	workers      map[int]*Worker
	nextID       int
	wg           sync.WaitGroup
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		inputChannel: make(chan string),
		workers:      make(map[int]*Worker),
		nextID:       0,
	}
}

func (pool *WorkerPool) AddWorker() {
	worker := &Worker{
		id:         pool.nextID,
		stopSignal: make(chan bool),
	}
	pool.workers[pool.nextID] = worker
	pool.nextID++
	pool.wg.Add(1)
	go worker.Start(pool.inputChannel, &pool.wg)
	fmt.Printf("Worker %d added\n", worker.id)
}

func (pool *WorkerPool) RemoveWorker(id int) {
	worker, exists := pool.workers[id]
	if !exists {
		fmt.Printf("Worker %d does not exist\n", id)
		return
	}
	close(worker.stopSignal)
	delete(pool.workers, id)
	fmt.Printf("Worker %d removed\n", id)
}

func (pool *WorkerPool) StopAll() {
	for id, worker := range pool.workers {
		close(worker.stopSignal)
		fmt.Printf("Worker %d is stopping\n", id)
	}
	pool.wg.Wait()
	fmt.Println("All workers stopped")
}

func main() {
	pool := NewWorkerPool()

	pool.AddWorker()
	pool.AddWorker()

	go func() {
		for i := 0; i < 5; i++ {
			pool.inputChannel <- fmt.Sprintf("Task-%d", i)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	time.Sleep(2 * time.Second)
	pool.AddWorker()

	time.Sleep(1 * time.Second)
	pool.RemoveWorker(0)

	go func() {
		for i := 5; i < 8; i++ {
			pool.inputChannel <- fmt.Sprintf("Task-%d", i)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	time.Sleep(2 * time.Second)
	pool.StopAll()
}
