package cmd

import "fmt"

var Version = "dev"

func RunVersion() {
	fmt.Printf("zoro %s\n", Version)
}
