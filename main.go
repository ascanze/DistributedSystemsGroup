package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func philosopher(id int, forks []chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	left := id
	right := (id + 1) % 5
	meal := 1

	for meal <= 3 {
		fmt.Printf("Philosopher: %d is thinking...\n", id)
		time.Sleep(time.Duration(randRange(200, 800)) * time.Millisecond)

		if id%2 == 0 {
			<-forks[left]
			<-forks[right]
		} else {
			<-forks[right]
			<-forks[left]
		}

		fmt.Printf("Philosopher: %d is eating his meal %d\n", id, meal)
		time.Sleep(time.Duration(randRange(200, 800)) * time.Millisecond)
		meal++

		forks[left] <- "done"
		forks[right] <- "done"
	}
	fmt.Printf("Philosopher: %d is done eating.\n", id)
}

func fork(forkChan chan string) {
	for true {
		forkChan <- "grant"
		<-forkChan
	}
}

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	var wg sync.WaitGroup
	wg.Add(5)

	forks := make([]chan string, 5)

	for i := 0; i < 5; i++ {
		forks[i] = make(chan string)
		go fork(forks[i])
	}

	for i := 0; i < 5; i++ {
		go philosopher(i, forks, &wg)
	}

	wg.Wait()
	fmt.Println("All philosophers have finished eating!")
}
