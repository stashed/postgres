// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !gccgo,!purego

package poly1305

import (
	"golang.org/x/sys/cpu"
)

// updateVX is an assembly implementation of Poly1305 that uses vector
// instructions. It must only be called if the vector facility (vx) is
// available.
//go:noescape
func updateVX(state *macState, msg []byte)

// poly1305vmsl is an assembly implementation of Poly1305 that uses vector
// instructions, including VMSL. It must only be called if the vector facility (vx) is
// available and if VMSL is supported.
//go:noescape
func poly1305vmsl(out *[16]byte, m *byte, mlen uint64, key *[32]byte)

func sum(out *[16]byte, m []byte, key *[32]byte) {
	if cpu.S390X.HasVX {
		var mPtr *byte
		if len(m) > 0 {
			mPtr = &m[0]
		}
	}

	tail := len(p) % len(h.buffer) // number of bytes to copy into buffer
	body := len(p) - tail          // number of bytes to process now
	if body > 0 {
		if cpu.S390X.HasVX {
			updateVX(&h.macState, p[:body])
		} else {
			updateGeneric(&h.macState, p[:body])
		}
	}
	h.offset = copy(h.buffer[:], p[body:]) // copy tail bytes - can be 0
	return nn, nil
}

func (h *mac) Sum(out *[TagSize]byte) {
	state := h.macState
	remainder := h.buffer[:h.offset]

	// Use the generic implementation if we have 2 or fewer blocks left
	// to sum. The vector implementation has a higher startup time.
	if cpu.S390X.HasVX && len(remainder) > 2*TagSize {
		updateVX(&state, remainder)
	} else if len(remainder) > 0 {
		updateGeneric(&state, remainder)
	}
	finalize(out, &state.h, &state.s)
}
