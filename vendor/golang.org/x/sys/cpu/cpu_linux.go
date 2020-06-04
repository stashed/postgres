// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !amd64,!amd64p32,!386

package cpu

func init() {
	if err := readHWCAP(); err != nil {
		return
	}
	doinit()
	Initialized = true
}
