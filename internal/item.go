package internal

import (
	"fmt"

	"github.com/dustin/go-humanize"
)

type item struct {
	keyType string

	key string
	val string

	err bool
}

func (i item) Title() string { return i.key }

func (i item) Description() string {
	if i.err {
		return "get error: " + i.val
	}
	valLen := len(i.val)
	return fmt.Sprintf("key: %d bytes, value: %d B (%s)", len(i.key), valLen, humanize.Bytes(uint64(valLen)))
}

func (i item) FilterValue() string { return i.key }
