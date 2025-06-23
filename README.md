# BatChan

[![CI](https://github.com/floatdrop/batchan/actions/workflows/ci.yaml/badge.svg)](https://github.com/floatdrop/batchan/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/floatdrop/batchan)](https://goreportcard.com/report/github.com/floatdrop/batchan)
[![Go Reference](https://pkg.go.dev/badge/github.com/floatdrop/batchan.svg)](https://pkg.go.dev/github.com/floatdrop/batchan)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Installation

```bash
go get github.com/floatdrop/batchan
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/floatdrop/batchan"
)

func main() {
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
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
