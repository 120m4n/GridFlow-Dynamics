// Package worker provides a configurable worker pool for processing messages.
package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/repository"
	"github.com/120m4n/GridFlow-Dynamics/service-worker-ps/internal/subscriber"
)

// Pool represents a worker pool that processes messages concurrently.
type Pool struct {
	numWorkers      int
	repo            repository.Repository
	messageChan     chan *subscriber.Message
	wg              sync.WaitGroup
	shutdownTimeout time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewPool creates a new worker pool.
func NewPool(numWorkers int, bufferSize int, repo repository.Repository, shutdownTimeout time.Duration) *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		numWorkers:      numWorkers,
		repo:            repo,
		messageChan:     make(chan *subscriber.Message, bufferSize),
		shutdownTimeout: shutdownTimeout,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Start initializes and starts all workers in the pool.
func (p *Pool) Start() {
	log.Printf("Starting worker pool with %d workers", p.numWorkers)

	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	log.Printf("Worker pool started successfully")
}

// worker is a goroutine that processes messages from the channel.
func (p *Pool) worker(id int) {
	defer p.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case msg, ok := <-p.messageChan:
			if !ok {
				log.Printf("Worker %d: message channel closed, shutting down", id)
				return
			}
			p.processMessage(id, msg)

		case <-p.ctx.Done():
			log.Printf("Worker %d: context cancelled, shutting down", id)
			return
		}
	}
}

// processMessage handles a single message by storing it in the repository.
func (p *Pool) processMessage(workerID int, msg *subscriber.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := &repository.InventarioData{
		GrupoTrabajo:       msg.GrupoTrabajo,
		NombreEmpleado:     msg.NombreEmpleado,
		Timestamp:          msg.Timestamp,
		Latitud:            msg.Coordenadas.Latitud,
		Longitud:           msg.Coordenadas.Longitud,
		CodigoODT:          msg.CodigoODT,
		Estado:             msg.Estado,
		PorcentajeProgreso: msg.PorcentajeProgreso,
		NivelBateria:       msg.NivelBateria,
	}

	if err := p.repo.Save(ctx, data); err != nil {
		log.Printf("Worker %d: failed to save message to repository: %v", workerID, err)
		return
	}

	log.Printf("Worker %d: successfully saved message from cuadrilla %s (ODT: %s, Estado: %s)",
		workerID, data.GrupoTrabajo, data.CodigoODT, data.Estado)
}

// Submit adds a message to the worker pool for processing.
// This method is non-blocking and returns immediately.
func (p *Pool) Submit(msg *subscriber.Message) {
	select {
	case p.messageChan <- msg:
		// Message submitted successfully
	case <-p.ctx.Done():
		log.Printf("Worker pool is shutting down, message from %s dropped", msg.GrupoTrabajo)
	default:
		log.Printf("Worker pool buffer full, message from %s dropped", msg.GrupoTrabajo)
	}
}

// Shutdown gracefully shuts down the worker pool.
// It waits for all workers to finish processing their current messages
// or until the shutdown timeout is reached.
func (p *Pool) Shutdown() error {
	log.Println("Initiating worker pool shutdown...")

	// Cancel context to signal workers to stop
	p.cancel()

	// Close the message channel (no new messages will be accepted)
	close(p.messageChan)

	// Wait for all workers to finish with a timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers shut down gracefully")
	case <-time.After(p.shutdownTimeout):
		log.Printf("Worker pool shutdown timed out after %v", p.shutdownTimeout)
	}

	return nil
}
