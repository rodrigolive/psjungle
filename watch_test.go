package main

import (
	"fmt"
	"time"
)

func testWatch() {
	fmt.Println("Testing watch functionality")
	for i := 0; i < 3; i++ {
		fmt.Printf("Iteration %d\n", i+1)
		time.Sleep(1 * time.Second)
	}
}
