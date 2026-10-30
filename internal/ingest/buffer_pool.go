package ingest

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/alexandrmotologa/microhog/internal/storage"
)

// BufferPool queues events in memory and writes them in batches to reduce SQLite disk operations.
type BufferPool struct {
	eventStore    *storage.EventStore
	ch            chan *storage.Event
	flushInterval time.Duration
	batchSize     int
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewBufferPool initializes a buffered ingestion queue.
func NewBufferPool(eventStore *storage.EventStore, capacity, batchSize int, flushInterval time.Duration) *BufferPool {
	ctx, cancel := context.WithCancel(context.Background())
	bp := &BufferPool{
		eventStore:    eventStore,
		ch:            make(chan *storage.Event, capacity),
		flushInterval: flushInterval,
		batchSize:     batchSize,
		ctx:           ctx,
		cancel:        cancel,
	}

	bp.wg.Add(1)
	go bp.worker()

	return bp
}

// Enqueue adds an event to the buffer. Drops if the channel is completely full to preserve memory.
func (bp *BufferPool) Enqueue(e *storage.Event) bool {
	select {
	case bp.ch <- e:
		return true
	default:
		log.Printf("[BufferPool] Warning: buffer channel full, dropping event %s", e.EventName)
		return false
	}
}

func (bp *BufferPool) worker() {
	defer bp.wg.Done()

	ticker := time.NewTicker(bp.flushInterval)
	defer ticker.Stop()

	batch := make([]*storage.Event, 0, bp.batchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := bp.eventStore.InsertBatch(batch); err != nil {
			log.Printf("[BufferPool] Error inserting event batch of %d: %v", len(batch), err)
		}
		batch = make([]*storage.Event, 0, bp.batchSize)
	}

	for {
		select {
		case <-bp.ctx.Done():
			// Drain remaining events from channel
			for {
				select {
				case e := <-bp.ch:
					batch = append(batch, e)
					if len(batch) >= bp.batchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}

		case e := <-bp.ch:
			batch = append(batch, e)
			if len(batch) >= bp.batchSize {
				flush()
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flush()
			}
		}
	}
}

// Stop signals the worker to finish processing and flush remaining events.
func (bp *BufferPool) Stop() {
	bp.cancel()
	bp.wg.Wait()
}
