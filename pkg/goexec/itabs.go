package goexec

import (
	"debug/elf"
	"fmt"
	"strings"
)

// TODO: override by user
const InterfaceName = "main.Greeter"

func findInterfaceImpls(ef *elf.File) (map[uint64]string, error) {
	implementations := map[uint64]string{}
	symbols, err := ef.Symbols()
	if err != nil {
		return nil, fmt.Errorf("accessing symbols table: %w", err)
	}
	for _, s := range symbols {
		if !strings.Contains(s.Name, "go:itab") {
			continue
		}
		if strings.HasSuffix(s.Name, InterfaceName) {
			implementations[s.Value] = s.Name
		}
	}
	return implementations, nil
}
