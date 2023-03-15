package process

import (
	"fmt"
	"github.com/mariomac/ebpf-go-interface-matching/pkg/goexec"
)

type Namer struct {
	Itabs map[uint64]goexec.ITabInfo
}

func (n *Namer) Do(in <-chan uint64, out chan<- string) {
	for addr := range in {
		itab, ok := n.Itabs[addr]
		if ok {
			out <- itab.InterfaceName + " implemented by " + itab.ImplementorName
		} else {
			out <- fmt.Sprintf("notfound:0x%x", addr)
		}
	}
}
