package batchan

import "time"

type option[T any] func(*config[T])

type config[T any] struct {
	timeout    time.Duration
	hasTimeout bool
	splitFunc  func(T, T) bool
}

func WithTimeout[T any](timeout time.Duration) option[T] {
	return func(cfg *config[T]) {
		cfg.timeout = timeout
		cfg.hasTimeout = true
	}
}

func WithSplitFunc[T any](splitFunc func(T, T) bool) option[T] {
	return func(cfg *config[T]) {
		cfg.splitFunc = splitFunc
	}
}

func timerOrNil(t *time.Timer) <-chan time.Time {
	if t != nil {
		return t.C
	}
	return nil
}

func noSplitFunc[T any](t1, t2 T) bool { return false }

func New[T any](in <-chan T, size int, opts ...option[T]) <-chan []T {
	cfg := &config[T]{
		splitFunc: noSplitFunc[T],
	}

	for _, opt := range opts {
		opt(cfg)
	}

	out := make(chan []T)

	go func() {
		defer close(out)

		var (
			currentBatch []T
			timer        *time.Timer
		)

		if cfg.hasTimeout {
			timer = time.NewTimer(cfg.timeout)
			defer timer.Stop()
		}

		flush := func() {
			if len(currentBatch) > 0 {
				out <- currentBatch
				currentBatch = nil
			}
			if cfg.hasTimeout {
				timer.Reset(cfg.timeout)
			}
		}

		for {
			select {
			case t, ok := <-in:
				if !ok {
					flush()
					return
				}

				if len(currentBatch) > 0 && cfg.splitFunc(currentBatch[len(currentBatch)-1], t) {
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
