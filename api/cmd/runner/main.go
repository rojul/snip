package main

import (
	"encoding/json"
	"os"
	"runtime"

	"github.com/rojul/snip/api/runner"
)

func main() {
	runtime.GOMAXPROCS(1)

	res := runner.Run(os.Stdin)
	json.NewEncoder(os.Stdout).Encode(res)
}
