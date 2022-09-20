// Copyright 2022 Unknwon. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package conf

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)
