package batchan_test

import (
	"fmt"

	"github.com/floatdrop/batchan"
)

func ExampleNew() {
	input := make(chan string, 5)
	batches := batchan.New(input, 3)

	go func() {
		inputs := []string{"A", "B", "C", "D", "E"}
		for _, v := range inputs {
			input <- v
		}
		close(input)
	}()

	for v := range batches {
		fmt.Println("Got:", v)
	}

	// Output:
	// Got: [A B C]
	// Got: [D E]
}
