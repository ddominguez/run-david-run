package main

import (
	"fmt"

	"github.com/ddominguez/run-david-run/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
