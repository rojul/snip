package main

import (
	"os"
	"runtime"

	"github.com/rojul/snip/api/runner"
)

func main() {
	runtime.GOMAXPROCS(1)

	runner.Run(os.Stdin, os.Stdout)
}
