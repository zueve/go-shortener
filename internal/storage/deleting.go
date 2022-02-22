package storage

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/zueve/go-shortener/pkg/logging"
)

type StorageExpected interface {
	DeleteByBatch(ctx context.Context, batch []Task) error
}

type Task struct {
	URL    string
	UserID string
}

type Deleter struct {
	workers []Worker
	stoping bool
	cancel  func()
	pushWg  *sync.WaitGroup
}

func NewDeleter(storage StorageExpected, batchSize int, workerNum int, pushPeriod time.Duration) (*Deleter, error) {
	var wg sync.WaitGroup
	workers := make([]Worker, workerNum)
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < workerNum; i++ {
		ch := make(chan Task)
		w := Worker{BatchSize: batchSize, Storage: storage, InpCh: ch}
		workers[i] = w
		go w.Loop(ctx, pushPeriod)
	}
	return &Deleter{workers: workers, stoping: false, cancel: cancel, pushWg: &wg}, nil
}

func (d *Deleter) Push(task Task) error {
	if d.stoping {
		return errors.New("cancelation in progress")
	}
	inx := rand.Intn(len(d.workers))
	d.pushWg.Add(1)
	go func(inp chan Task) {
		defer d.pushWg.Done()
		inp <- task
	}(d.workers[inx].InpCh)
	return nil
}

func (d *Deleter) Shutdown() error {
	d.stoping = true
	d.pushWg.Wait()
	d.cancel()
	return nil
}

type Worker struct {
	BatchSize int
	InpCh     chan Task
	Storage   StorageExpected
	batch     []Task
}

func (w *Worker) Loop(ctx context.Context, period time.Duration) {
	timer := time.NewTicker(period)
	for {
		select {
		case task := <-w.InpCh:
			w.batch = append(w.batch, task)
			// run by batch size
			if len(w.batch) == w.BatchSize {
				w.Run(ctx)
			}
		case <-timer.C:
			// run by timer
			w.Run(ctx)
		case <-ctx.Done():
			w.Run(ctx)
			return
		}
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.log(ctx).Debug().Msgf("Run with batch %d", len(w.batch))
	if err := w.Storage.DeleteByBatch(context.Background(), w.batch); err != nil {
		w.log(ctx).Err(err).Msg("error on DeleteByBatch")
	}
	w.batch = make([]Task, 0)
}

func (w *Worker) log(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().
		Str(logging.Source, "Deleter").
		Str(logging.Layer, "Storage").
		Logger()

	return &logger
}
