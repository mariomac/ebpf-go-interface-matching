package process

import "fmt"

type Namer struct {
	Itabs map[uint64]string
}

func (n *Namer) Do(in <-chan uint64, out chan<- string) {
	for addr := range in {
		name, ok := n.Itabs[addr]
		if ok {
			out <- name
		} else {
			out <- fmt.Sprintf("notfound:0x%x", addr)
		}
	}
}
