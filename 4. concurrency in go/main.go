package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.Tick(1 * time.Second)

	for i := 0; i < 50; i++ {
		tickTime := <-ticker
		fmt.Println("Tick at", tickTime)
	}
}
