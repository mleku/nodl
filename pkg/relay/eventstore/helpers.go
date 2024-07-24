package eventstore

import (
	"bytes"
)

func isOlder(prev, next EV) bool {
	return prev.CreatedAt.I64() < next.CreatedAt.I64() ||
		(prev.CreatedAt == next.CreatedAt && bytes.Compare(prev.ID, next.ID) < 0)
}
