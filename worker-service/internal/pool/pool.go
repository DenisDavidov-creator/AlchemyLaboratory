package pool

import "sync"

type Pool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

func New(size int) *Pool {
	p := &Pool{
		tasks: make(chan func()),
	}

	p.wg.Add(size)
	for range size {
		go func() {
			defer p.wg.Done()
			for task := range p.tasks {
				task()
			}
		}()
	}
	return p
}

func (p *Pool) Submit(task func()) {
	p.tasks <- task
}

func (p *Pool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}
