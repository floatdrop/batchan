package batchan_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/floatdrop/batchan"
)

func sendIntsToChan(data []int, delay time.Duration) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range data {
			ch <- v
			time.Sleep(delay)
		}
	}()
	return ch
}

func collectBatches[T any](out <-chan []T) [][]T {
	var result [][]T
	for batch := range out {
		result = append(result, batch)
	}
	return result
}

func TestBatchingFlushOnTimeout(t *testing.T) {
	in := sendIntsToChan([]int{1, 2}, 150*time.Millisecond) // delay > timeout
	out := batchan.New(in, 5, batchan.WithTimeout[int](100*time.Millisecond))

	got := collectBatches(out)

	expected := [][]int{{1}, {2}}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchingFlushOnSizeOrTimeout(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3, 4}, 50*time.Millisecond)
	out := batchan.New(in, 2, batchan.WithTimeout[int](200*time.Millisecond))

	got := collectBatches(out)
	expected := [][]int{{1, 2}, {3, 4}}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestFlushTimeoutMultiple(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3}, 300*time.Millisecond)
	out := batchan.New(in, 10, batchan.WithTimeout[int](200*time.Millisecond)) // small timeout, large batch size

	got := collectBatches(out)
	expected := [][]int{{1}, {2}, {3}}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestTimeoutResetsAfterFlush(t *testing.T) {
	// This test checks that the timer is correctly reset after each flush
	in := make(chan int)
	go func() {
		defer close(in)
		in <- 1
		time.Sleep(150 * time.Millisecond)
		in <- 2
		time.Sleep(150 * time.Millisecond)
		in <- 3
	}()

	out := batchan.New(in, 2, batchan.WithTimeout[int](100*time.Millisecond))

	got := collectBatches(out)
	expected := [][]int{{1}, {2}, {3}}

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchingBasic(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3, 4, 5, 6}, time.Microsecond)
	out := batchan.New(in, 2)

	expected := [][]int{{1, 2}, {3, 4}, {5, 6}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchingWithRemainder(t *testing.T) {
	in := sendIntsToChan([]int{10, 20, 30, 40, 50}, time.Microsecond)
	out := batchan.New(in, 2)

	expected := [][]int{{10, 20}, {30, 40}, {50}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestEmptyInput(t *testing.T) {
	in := sendIntsToChan([]int{}, time.Microsecond)
	out := batchan.New(in, 3)

	got := collectBatches(out)

	if len(got) != 0 {
		t.Errorf("Expected empty output, got %v", got)
	}
}

func TestBatchSizeOne(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3}, time.Microsecond)
	out := batchan.New(in, 1)

	expected := [][]int{{1}, {2}, {3}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchSizeLargerThanInput(t *testing.T) {
	in := sendIntsToChan([]int{42, 99}, time.Microsecond)
	out := batchan.New(in, 5)

	expected := [][]int{{42, 99}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestSplitFunc(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3, 5, 6}, time.Microsecond)
	out := batchan.New(in, 5, batchan.WithSplitFunc(func(i1, i2 int) bool { return i2-i1 > 1 }))

	expected := [][]int{{1, 2, 3}, {5, 6}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

// Optional: Test that the output channel closes properly
func TestOutputChannelClosure(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3}, time.Microsecond)
	out := batchan.New(in, 2)

	timeout := time.After(1 * time.Second)
	for {
		select {
		case _, ok := <-out:
			if !ok {
				return // success
			}
		case <-timeout:
			t.Fatal("Output channel did not close")
		}
	}
}
