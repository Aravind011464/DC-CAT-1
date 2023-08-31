package main

import (
	"fmt"
	"time"
	"sync"
	"math/rand"
) 

type Process struct {
	id       int
	accounts [3]int
	channel  chan Transaction
	snapshot Snapshot
	mu       sync.Mutex
}

type Transaction struct {
	amount int
	from   int
}

type Snapshot struct {
	accounts [3]int
	channel  []Transaction
}

var processes []*Process
var wg sync.WaitGroup

func main() {
	var n int
	fmt.Print("Enter the number of processes: ")
	fmt.Scan(&n)

	for i := 0; i < n; i++ {
		process := &Process{
			id: i + 1,
			channel: make(chan Transaction, 100),
		}
		processes = append(processes, process) 
	}

	for _, process := range processes {
		for i := 0; i < 3; i++ {
			process.accounts[i] = 1000000
		}
	}

	totalAmount := getTotalAmount()
	fmt.Printf("Total amount in the system: %d\n", totalAmount)

	for _, process := range processes {
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			for {
				transaction := generateTransaction(p)
				p.mu.Lock()
				p.accounts[transaction.from] -= transaction.amount
				p.accounts[p.id-1] += transaction.amount
				p.mu.Unlock()

				p.channel <- transaction

				if rand.Intn(n*15) == 0 {
					p.snapshotState()
				}

				time.Sleep(time.Millisecond * 100) 
			}
		}(process)
	}

	for _, process := range processes {
		close(process.channel)
	}
}

func (p *Process) snapshotState() {
	p.mu.Lock()
	defer p.mu.Unlock()

	var snapshot Snapshot
	snapshot.accounts = p.accounts

	p.mu.Lock()
	channelSlice := make([]Transaction, len(p.channel))
	close(p.channel)
	idx := 0
	for transaction := range p.channel {
		channelSlice[idx] = transaction
		idx++
	}
	p.channel = make(chan Transaction, 100) 
	p.mu.Unlock()

	snapshot.channel = channelSlice

	p.snapshot = snapshot

	snapshotTotal := getTotalAmountFromSnapshot(p.snapshot)
	fmt.Printf("Snapshot for Process %d: Total amount = %d\n", p.id, snapshotTotal)
}

func getTotalAmount() int {
	total := 0
	for _, process := range processes {
		process.mu.Lock()
		for _, amount := range process.accounts {
			total += amount
		}
		process.mu.Unlock()
	}
	return total
}

func getTotalAmountFromSnapshot(snapshot Snapshot) int {
	total := 0
	for _, amount := range snapshot.accounts {
		total += amount
	}
	return total
}

func generateTransaction(p *Process) Transaction {
	amount := rand.Intn(1000) + 1
	from := rand.Intn(3)
	return Transaction{
		amount: amount,
		from:   from,
	}
}
