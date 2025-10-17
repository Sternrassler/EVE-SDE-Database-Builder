package worker

import (
	"context"
	"sync"
)

// Job repräsentiert eine zu verarbeitende Aufgabe im Worker Pool.
//
// Ein Job besteht aus einer eindeutigen ID und einer auszuführenden Funktion.
// Die Funktion erhält einen Context für Cancellation und gibt ein Ergebnis
// sowie einen optionalen Fehler zurück.
//
// Beispiel:
//
//	job := worker.Job{
//	    ID: "parse-types",
//	    Fn: func(ctx context.Context) (interface{}, error) {
//	        return parseFile(ctx, "types.jsonl")
//	    },
//	}
type Job struct {
	ID string                                      // Eindeutige Job-Identifikation
	Fn func(context.Context) (interface{}, error) // Auszuführende Funktion
}

// Result repräsentiert das Ergebnis einer Job-Ausführung.
//
// Result enthält die Job-ID, die zurückgegebenen Daten und einen
// optionalen Fehler. Results werden vom Pool gesammelt und können
// über Wait() abgerufen werden.
type Result struct {
	JobID string      // ID des ausgeführten Jobs
	Data  interface{} // Rückgabewert der Job-Funktion
	Err   error       // Fehler (falls aufgetreten)
}

// Pool ist ein Worker Pool für parallele Job-Verarbeitung.
//
// Der Pool verwaltet eine konfigurierbare Anzahl von Worker-Goroutines,
// die Jobs aus einer Warteschlange verarbeiten. Buffered Channels sorgen
// für Backpressure und vermeiden Blockierung.
//
// Ein Pool wird mit NewPool() erstellt, mit Start() gestartet und mit
// Wait() auf Completion gewartet.
//
// Beispiel:
//
//	pool := worker.NewPool(4)
//	pool.Start(ctx)
//	pool.Submit(job)
//	results, errors := pool.Wait()
type Pool struct {
	workers int
	jobs    chan Job
	results chan Result
	wg      sync.WaitGroup
}

// NewPool erstellt einen neuen Worker Pool mit der angegebenen Anzahl von Workers.
//
// Die Anzahl der Workers bestimmt, wie viele Jobs parallel verarbeitet werden können.
// Bei workers <= 0 wird automatisch 1 Worker verwendet.
//
// Der Pool ist nach Erstellung noch nicht aktiv. Start() muss aufgerufen werden,
// um die Worker-Goroutines zu starten.
//
// Beispiel:
//
//	pool := worker.NewPool(runtime.NumCPU()) // Optimal für CPU-bound Tasks
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

// Start startet die Worker Goroutines.
//
// Der übergebene Context wird von allen Workern respektiert. Bei ctx.Done()
// beenden sich die Worker nach Abschluss des aktuellen Jobs gracefully.
//
// Start sollte nur einmal pro Pool aufgerufen werden, bevor Jobs submitted werden.
//
// Beispiel:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//	pool.Start(ctx)
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

// Submit fügt einen Job zur Warteschlange hinzu.
//
// Jobs werden asynchron verarbeitet. Submit blockiert, wenn der interne
// Job-Channel voll ist (Backpressure).
//
// Submit sollte nicht nach Wait() aufgerufen werden.
//
// Beispiel:
//
//	pool.Submit(worker.Job{
//	    ID: "task-1",
//	    Fn: func(ctx context.Context) (interface{}, error) {
//	        return processTask(ctx)
//	    },
//	})
func (p *Pool) Submit(job Job) {
	p.jobs <- job
}

// Wait wartet bis alle Worker fertig sind und gibt Results und Errors zurück.
//
// Wait schließt den Job-Channel, wartet auf alle Worker und sammelt alle Results.
// Fehler werden zusätzlich in einem separaten Slice zurückgegeben.
//
// Wait sollte nur einmal pro Pool aufgerufen werden, nachdem alle Jobs submitted wurden.
//
// Rückgabewerte:
//   - []Result: Alle Job-Results (inkl. fehlgeschlagener Jobs)
//   - []error: Nur die Fehler aus fehlgeschlagenen Jobs
//
// Beispiel:
//
//	results, errors := pool.Wait()
//	if len(errors) > 0 {
//	    log.Printf("Fehler bei %d/%d Jobs", len(errors), len(results))
//	}
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
