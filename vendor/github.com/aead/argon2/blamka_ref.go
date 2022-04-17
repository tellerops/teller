// Copyright (c) 2017 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

// +build !amd64 appengine gccgo

package argon2

func processBlock(out, in1, in2 *block) {
	processBlockGeneric(out, in1, in2, false)
}

func processBlockXOR(out, in1, in2 *block) {
	processBlockGeneric(out, in1, in2, true)
}
