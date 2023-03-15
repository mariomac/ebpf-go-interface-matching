// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ebpf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/mariomac/ebpf-go-interface-matching/pkg/goexec"
	"golang.org/x/exp/slog"
)

//go:generate bpf2go -cc $BPF_CLANG -cflags $BPF_CFLAGS -target amd64,arm64 bpf ../../bpf/greeter.c -- -I../../bpf/headers

type InstrumentedGreeter struct {
	bpfObjects   bpfObjects
	uprobe       link.Link
	eventsReader *ringbuf.Reader
}

// Instrument the executable passed as path and insert probes in the provided offsets, so the
// returned InstrumentedGreeter instance will listen and forward traces for each HTTP invocation.
func Instrument(offsets *goexec.Offsets) (*InstrumentedGreeter, error) {
	// Instead of the executable file in the disk, we pass the /proc/<pid>/exec
	// to allow loading it from different container/pods in containerized environments
	exe, err := link.OpenExecutable(offsets.FileInfo.ProExeLinkPath)
	if err != nil {
		return nil, fmt.Errorf("opening executable file %q: %w",
			offsets.FileInfo.ProExeLinkPath, err)
	}

	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("removing memlock: %w", err)
	}

	spec, err := loadBpf()
	if err != nil {
		return nil, fmt.Errorf("loading BPF data: %w", err)
	}

	h := InstrumentedGreeter{}
	// Load BPF programs
	if err := spec.LoadAndAssign(&h.bpfObjects, nil); err != nil {
		return nil, fmt.Errorf("loading and assigning BPF objects: %w", err)
	}
	// Attach BPF programs as start and return probes
	up, err := exe.Uprobe("", h.bpfObjects.UprobeGreet, &link.UprobeOptions{
		Address: offsets.Func.Start,
	})
	if err != nil {
		return nil, fmt.Errorf("setting uprobe: %w", err)
	}
	h.uprobe = up

	// BPF will send each measured trace via Ring Buffer, so we listen for them from the
	// user space.
	rd, err := ringbuf.NewReader(h.bpfObjects.Greets)
	if err != nil {
		return nil, fmt.Errorf("creating perf reader: %w", err)
	}
	h.eventsReader = rd

	return &h, nil
}

func (h *InstrumentedGreeter) Run(eventsChan chan<- uint64) {
	logger := slog.With("name", "InstrumentedGreeter")
	for {
		logger.Debug("starting to read perf buffer")
		record, err := h.eventsReader.Read()
		logger.Debug("received event")
		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) {
				return
			}
			logger.Error("error reading from perf reader", err)
			continue
		}

		var greetImplItab uint64
		if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &greetImplItab); err != nil {
			logger.Error("error parsing perf event", err)
			continue
		}

		eventsChan <- greetImplItab
	}
}

func (h *InstrumentedGreeter) Close() {
	slog.With("name", "InstrumentedGreeter").Info("closing instrumenter")
	if h.eventsReader != nil {
		h.eventsReader.Close()
	}

	if h.uprobe != nil {
		h.uprobe.Close()
	}
	h.bpfObjects.Close()
}
