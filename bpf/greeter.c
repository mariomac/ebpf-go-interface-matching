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

#include "utils.h"

char __license[] SEC("license") = "Dual MIT/GPL";

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24);
} greets SEC(".maps");

SEC("uprobe/Greet")
int uprobe_greet(struct pt_regs *ctx) {
    bpf_printk("=== uprobe/Greet === ");
    // Param 1, pointer to the interface implementer itab
    // Param 2, pointer to the actual interface variable. We will use it when we want to access concrete data.
    void *itab = GO_PARAM1(ctx);
    bpf_printk("itab %lx", itab);

    // We just submit the pointer to the itab, which will be matched with the type name at the user space
    void **submit_itab = bpf_ringbuf_reserve(&greets, sizeof(void*), 0);
    if (!submit_itab) {
        bpf_printk("can't reserve space in the ringbuffer");
        return 0;
    }

    *submit_itab = itab;
    bpf_ringbuf_submit(submit_itab, 0);

    return 0;
}
