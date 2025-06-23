package batchan

func New[T any](in <-chan T, size int) <-chan []T {
	out := make(chan []T)

	go func() {
		defer close(out)

		var currentBatch []T

		flush := func() {
			if len(currentBatch) > 0 {
				out <- currentBatch
				currentBatch = nil
			}
		}

		for t := range in {
			if len(currentBatch) >= size {
				flush()
			}
			currentBatch = append(currentBatch, t)
		}

		flush()
	}()

	return out
}
