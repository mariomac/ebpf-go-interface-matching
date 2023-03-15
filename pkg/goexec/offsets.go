// Package goexec helps analyzing Go executables
package goexec

import "fmt"

type Offsets struct {
	FileInfo FileInfo
	Func     FuncOffsets
	Itabs    map[uint64]ITabInfo
}

type ITabInfo struct {
	InterfaceName   string
	ImplementorName string
}

type FuncOffsets struct {
	Start uint64
}

type FieldOffsets map[string]any

// InspectOffsets gets the memory addresses/offsets of the instrumenting function, as well as the required
// parameters fields to be read from the eBPF code
func InspectOffsets(execFile, funcName string) (Offsets, error) {
	// Analyse executable ELF file and find instrumentation points
	execElf, err := findExecELF(execFile)
	if err != nil {
		return Offsets{}, fmt.Errorf("looking for %s executable ELF: %w", execFile, err)
	}
	defer execElf.ELF.Close()

	// check the function instrumentation points
	funcOffsets, err := instrumentationPoints(execElf.ELF, funcName)
	if err != nil {
		return Offsets{}, fmt.Errorf("searching for instrumentation points for func %s in file %s: %w",
			funcName, execFile, err)
	}

	implementations, err := findInterfaceImpls(execElf.ELF)
	if err != nil {
		return Offsets{}, fmt.Errorf("checking interface implementations in file %s: %w", execFile, err)
	}

	return Offsets{
		FileInfo: execElf,
		Func:     funcOffsets,
		Itabs:    implementations,
	}, nil
}
