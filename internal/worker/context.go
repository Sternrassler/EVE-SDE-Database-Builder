package worker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler erstellt einen Context der bei SIGINT/SIGTERM cancelled wird.
// Dies ermöglicht Graceful Shutdown bei Ctrl+C oder Termination Signals.
//
// Der zurückgegebene Context sollte an Pool.Start() und Orchestrator.ImportAll()
// weitergegeben werden, um sicherzustellen dass alle Worker gracefully beendet werden.
//
// Beispiel:
//
//	ctx := SetupSignalHandler()
//	pool := NewPool(4)
//	pool.Start(ctx)
//	// Workers werden bei SIGINT/SIGTERM automatisch beendet
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	return ctx
}
