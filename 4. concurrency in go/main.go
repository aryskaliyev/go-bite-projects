package main

import (
	"fmt"
	"time"
)

func reallyLongCalculation(value interface{}) {
	time.Sleep(2 * time.Second)
	fmt.Println("[+] Really long calculation executed!")
}

func main() {
	var value interface{}
	select {
	case <-done:
		return
	case value = <-valueStream:
	}

	result := reallyLongCalculation(value)

	select {
	case <-done:
		return
	case resultStream<- result:
	}
}
