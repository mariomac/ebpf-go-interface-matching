package goexec

import (
	"debug/elf"
	"fmt"
	"strings"
)

// TODO: override by user
const InterfaceName = "main.Greeter"

func findInterfaceImpls(ef *elf.File) (map[uint64]ITabInfo, error) {
	implementations := map[uint64]ITabInfo{}
	symbols, err := ef.Symbols()
	if err != nil {
		return nil, fmt.Errorf("accessing symbols table: %w", err)
	}
	for _, s := range symbols {
		// Name is in format: go:itab.*net/http.response,net/http.ResponseWriter
		if !strings.Contains(s.Name, "go:itab.") {
			continue
		}
		parts := strings.Split(s.Name[len("go:itab."):], ",")
		if len(parts) < 2 {
			continue
		}
		if parts[1] == InterfaceName {
			implementations[s.Value] = ITabInfo{
				InterfaceName:   parts[1],
				ImplementorName: parts[0],
			}
		}
	}
	return implementations, nil
}
