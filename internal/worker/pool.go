package worker

import (
	"context"
	"sync"
)

// Job repräsentiert eine zu verarbeitende Aufgabe
type Job struct {
	ID string
	Fn func(context.Context) (interface{}, error)
}

// Result repräsentiert das Ergebnis einer Job-Ausführung
type Result struct {
	JobID string
	Data  interface{}
	Err   error
}

// Pool ist ein Worker Pool für parallele Job-Verarbeitung
type Pool struct {
	workers int
	jobs    chan Job
	results chan Result
	wg      sync.WaitGroup
}

// NewPool erstellt einen neuen Worker Pool mit der angegebenen Anzahl von Workers
func NewPool(workers int) *Pool {
	if workers <= 0 {
		workers = 1
	}

	return &Pool{
		workers: workers,
		jobs:    make(chan Job, workers*2),
		results: make(chan Result, 100), // Large buffer to prevent blocking
	}
}

// Start startet die Worker Goroutines
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
}

// worker verarbeitet Jobs aus dem Channel
func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			// Context wurde gecancelt, Worker stoppt
			return

		case job, ok := <-p.jobs:
			if !ok {
				// Channel wurde geschlossen, Worker stoppt
				return
			}

			// Job ausführen
			data, err := job.Fn(ctx)

			// Result in Channel schreiben (non-blocking bei Context-Cancel)
			select {
			case p.results <- Result{JobID: job.ID, Data: data, Err: err}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Submit fügt einen Job zur Warteschlange hinzu
func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

// Wait wartet bis alle Worker fertig sind und gibt Results und Errors zurück
func (p *Pool) Wait() ([]Result, []error) {
	close(p.jobs)
	p.wg.Wait()
	close(p.results)

	var results []Result
	var errors []error

	for result := range p.results {
		results = append(results, result)
		if result.Err != nil {
			errors = append(errors, result.Err)
		}
	}

	return results, errors
}
