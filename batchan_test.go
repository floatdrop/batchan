package batchan_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/floatdrop/batchan"
)

func sendIntsToChan(data []int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range data {
			ch <- v
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

func TestBatchingBasic(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3, 4, 5, 6})
	out := batchan.New(in, 2)

	expected := [][]int{{1, 2}, {3, 4}, {5, 6}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchingWithRemainder(t *testing.T) {
	in := sendIntsToChan([]int{10, 20, 30, 40, 50})
	out := batchan.New(in, 2)

	expected := [][]int{{10, 20}, {30, 40}, {50}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestEmptyInput(t *testing.T) {
	in := sendIntsToChan([]int{})
	out := batchan.New(in, 3)

	got := collectBatches(out)

	if len(got) != 0 {
		t.Errorf("Expected empty output, got %v", got)
	}
}

func TestBatchSizeOne(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3})
	out := batchan.New(in, 1)

	expected := [][]int{{1}, {2}, {3}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestBatchSizeLargerThanInput(t *testing.T) {
	in := sendIntsToChan([]int{42, 99})
	out := batchan.New(in, 5)

	expected := [][]int{{42, 99}}
	got := collectBatches(out)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

// Optional: Test that the output channel closes properly
func TestOutputChannelClosure(t *testing.T) {
	in := sendIntsToChan([]int{1, 2, 3})
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
