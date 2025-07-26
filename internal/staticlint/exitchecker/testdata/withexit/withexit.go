package main

import (
	"fmt"
	"os"
)

func main() {
	if 2*2 == 4 {
		fmt.Println("There is exit call")
		os.Exit(1) // want "os.Exit call in main function"
	}
}
