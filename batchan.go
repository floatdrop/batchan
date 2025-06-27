package batchan

import (
	"context"
	"time"
)

// Option defines a configuration option for the batcher.
type Option func(*config)

type config struct {
	ctx              context.Context
	timeout          time.Duration
	hasTimeout       bool
	outputBufferSize int
}

// WithTimeout sets the maximum duration to wait before emitting a batch,
// even if it hasn't reached the desired size.
func WithTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.timeout = timeout
		cfg.hasTimeout = true
	}
}

// WithOutputBufferSize sets the size of the output channel buffer.
// By default output channel is buffered with size 1.
func WithOutputBufferSize(size int) Option {
	return func(cfg *config) {
		cfg.outputBufferSize = size
	}
}

// WithContext sets a context for the batcher's lifetime. If the context
// is canceled, the batching process will stop and flush any remaining items.
func WithContext(ctx context.Context) Option {
	return func(cfg *config) {
		cfg.ctx = ctx
	}
}

// timerOrNil returns the timer channel if the timer exists, or nil otherwise.
func timerOrNil(t *time.Timer) <-chan time.Time {
	if t != nil {
		return t.C
	}
	return nil
}

// defaultSplitFunc is a default split function that never triggers a batch split.
func defaultSplitFunc[T any](t1, t2 T) bool { return false }

// New creates a batching channel that emits slices of type T.
// A new batch is emitted when the batch reaches the specified size or the optional timeout expires.
// Batching stops when the input channel is closed or the context is canceled.
func New[T any](in <-chan T, size int, opts ...Option) <-chan []T {
	return NewWithSplit(in, size, defaultSplitFunc, opts...)
}

// NewWithSplit creates a batching channel with an optional split function.
// The splitFunc is called with the last item of the current batch and the next item;
// if it returns true, the current batch is flushed before adding the new item.
func NewWithSplit[T any](in <-chan T, size int, splitFunc func(T, T) bool, opts ...Option) <-chan []T {
	cfg := &config{
		ctx:              context.Background(),
		outputBufferSize: 1,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	out := make(chan []T, cfg.outputBufferSize)

	go func() {
		defer close(out)

		var (
			currentBatch []T = make([]T, 0, size)
			timer        *time.Timer
		)

		if cfg.hasTimeout {
			timer = time.NewTimer(cfg.timeout)
			defer timer.Stop()
		}

		flush := func() {
			if len(currentBatch) > 0 {
				out <- currentBatch
				currentBatch = make([]T, 0, size)
			}
			if cfg.hasTimeout {
				timer.Reset(cfg.timeout)
			}
		}

		for {
			select {
			case <-cfg.ctx.Done():
				flush()
				return
			case t, ok := <-in:
				if !ok {
					flush()
					return
				}

				if len(currentBatch) > 0 && splitFunc(currentBatch[len(currentBatch)-1], t) {
					flush()
				}

				currentBatch = append(currentBatch, t)
				if len(currentBatch) >= size {
					flush()
				}
			case <-timerOrNil(timer):
				flush()
			}
		}
	}()

	return out
}
