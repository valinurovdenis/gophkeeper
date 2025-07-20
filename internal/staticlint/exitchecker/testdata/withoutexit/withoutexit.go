package main

import (
	"fmt"
	"os"
)

func fn() int {
	os.Exit(1)
	return 4
}

func main() {
	if fn() == 2*2 {
		fmt.Println("There isn't exit call")
	}
}
