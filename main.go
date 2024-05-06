package main

import (
	"fmt"
	"os"

	"github.com/s-nadesh/pydeohub/videohub"
)

func main() {
	ip := os.Args[1]

	fmt.Println("IP: ", ip)

	vh := videohub.NewVideohub(ip)
	// Now you can use methods of the Videohub struct, like vh.Route(), vh.InputLabel(), etc.

	// Use vh to perform some action, for example:
	vh.Route(1, 2) // Route output 1 to input 2
}
