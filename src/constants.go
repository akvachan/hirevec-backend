// Copyright (c) 2026 Arsenii Kvachan. All Rights Reserved. MIT License.

package hirevec

import "time"

const (
	bit      int64 = 1
	kilobyte int64 = bit * 1024
	megabyte int64 = kilobyte * 1024
)

const (
	pageSizeDefaultLimit = 50
	pageSizeMaxLimit     = 100
)

const maxBytesHandler = 1 * megabyte

const (
	ReadTimeout = 2 * time.Second
	WriteTimout = 2 * time.Second
)
