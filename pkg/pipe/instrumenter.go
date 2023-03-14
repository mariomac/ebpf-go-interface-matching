package pipe

import (
	"context"
	"fmt"

	"github.com/mariomac/ebpf-go-interface-matching/pkg/export"
	"github.com/mariomac/ebpf-go-interface-matching/pkg/process"

	"github.com/mariomac/ebpf-go-interface-matching/pkg/ebpf"

	"github.com/mariomac/ebpf-go-interface-matching/pkg/goexec"

	"github.com/mariomac/pipes/pkg/node"
)

type Graph struct {
	startNode *node.Start[uint64]
}

// builder with injectable instantiators for unit testing
type graphBuilder struct {
	config  *Config
	svcName string
}

// Build instantiates the whole instrumentation --> processing --> submit
// pipeline graph and returns it as a startable item
func Build(config *Config) (Graph, error) {
	if err := config.Validate(); err != nil {
		return Graph{}, fmt.Errorf("validating configuration: %w", err)
	}

	return (&graphBuilder{
		config: config,
	}).buildGraph()
}

func (gb *graphBuilder) buildGraph() (Graph, error) {
	offsetsInfo, err := goexec.InspectOffsets(gb.config.Exec, gb.config.FuncName)
	if err != nil {
		return Graph{}, fmt.Errorf("inspecting executable: %w", err)
	}
	gb.svcName = offsetsInfo.FileInfo.CmdExePath

	instrumentedServe, err := ebpf.Instrument(&offsetsInfo)
	if err != nil {
		return Graph{}, fmt.Errorf("instrumenting executable: %w", err)
	}
	// Build and connect the nodes of the processing pipeline
	//   greeter --> namer --> Printer
	tracer := node.AsStart(instrumentedServe.Run)
	namerImpl := process.Namer{Itabs: offsetsInfo.Itabs}
	namer := node.AsMiddle(namerImpl.Do)
	printer := node.AsTerminal(export.Printer)

	tracer.SendsTo(namer)
	namer.SendsTo(printer)

	return Graph{startNode: tracer}, nil
}

// Start the instrumentation --> processing --> submit pipeline
func (p *Graph) Start(ctx context.Context) {
	p.startNode.StartCtx(ctx)
}
