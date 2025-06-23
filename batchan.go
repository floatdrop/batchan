package batchan

import "time"

type option func(*config)

type config struct {
	timeout    time.Duration
	hasTimeout bool
}

func WithTimeout(timeout time.Duration) option {
	return func(cfg *config) {
		cfg.timeout = timeout
		cfg.hasTimeout = true
	}
}

func timerOrNil(t *time.Timer) <-chan time.Time {
	if t != nil {
		return t.C
	}
	return nil
}

func New[T any](in <-chan T, size int, opts ...option) <-chan []T {
	cfg := &config{}

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
