package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func philosopher(id int, forksReceive []chan string, forksSend []chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	left := id
	right := (id + 1) % 5
	meal := 1

	for meal <= 3 {
		fmt.Printf("Philosopher: %d is thinking...\n", id)
		time.Sleep(time.Duration(randRange(200, 800)) * time.Millisecond)

		if id%2 == 0 {
			<-forksReceive[left]
			<-forksReceive[right]
		} else {
			<-forksReceive[right]
			<-forksReceive[left]
		}

		fmt.Printf("Philosopher: %d is eating his meal %d\n", id, meal)
		time.Sleep(time.Duration(randRange(200, 800)) * time.Millisecond)
		meal++

		forksSend[left] <- "done"
		forksSend[right] <- "done"
	}
	fmt.Printf("Philosopher: %d is done eating.\n", id)
}

func fork(forkReceive chan string, forkSend chan string) {
	for true {
		forkReceive <- "grant"
		<-forkSend
	}
}

func randRange(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	var wg sync.WaitGroup
	wg.Add(5)

	forksReceive := make([]chan string, 5)
	forksSend := make([]chan string, 5)

	for i := 0; i < 5; i++ {
		forksReceive[i] = make(chan string)
		forksSend[i] = make(chan string)
		go fork(forksReceive[i], forksSend[i])
	}

	for i := 0; i < 5; i++ {
		go philosopher(i, forksReceive, forksSend, &wg)
	}

	wg.Wait()
	fmt.Println("All philosophers have finished eating!")
}
