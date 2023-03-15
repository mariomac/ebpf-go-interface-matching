package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Toy program used to check the interface type instrumentation.

type Greeter interface {
	Greet(name string)
}

type English struct{}

func (e *English) Greet(name string) {
	fmt.Println("hello", name)
}

type French struct{}

func (e *French) Greet(name string) {
	fmt.Println("bonjour", name)
}

type Japanese struct{}

func (e *Japanese) Greet(name string) {
	fmt.Println("今日は", name)
}

// DispatchMessage is the function. It receives a Greeter interface and, from eBPF,
// we want to know which concrete implementation type it is: French, Japanese or English.
//
//go:noinline
func DispatchMessage(greeter Greeter, name string) {
	greeter.Greet(name)
}

func main() {
	names := []string{"James", "坂本", "Amélie", "Pedro", "Antoine"}
	greeters := []Greeter{&English{}, &French{}, &Japanese{}}
	for {
		DispatchMessage(
			greeters[rand.Intn(len(greeters))],
			names[rand.Intn(len(names))])

		time.Sleep(time.Second)
	}
}
