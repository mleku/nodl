package main

import (
	"fmt"

	"github.com/mleku/nodl/pkg/interrupt"
)

func main() {
	interrupt.AddHandler(func() {
		fmt.Println("\rIT'S THE END OF THE WORLD!")
	})
	<-interrupt.HandlersDone
}
