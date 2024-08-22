package priority

import (
	"nostr.mleku.dev/codec/event"
	"git.replicatr.dev/pkg/relay/eventstore/badger/keys/serial"
)

type QueryEvent struct {
	*event.T
	Ser   *serial.T
	Query int
}

type Queue []*QueryEvent

func (pq *Queue) Len() int           { return len(*pq) }
func (pq *Queue) Less(i, j int) bool { return (*pq)[i].CreatedAt.I64() > (*pq)[j].CreatedAt.I64() }
func (pq *Queue) Swap(i, j int)      { (*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i] }
func (pq *Queue) Push(x any)         { *pq = append(*pq, x.(*QueryEvent)) }

func (pq *Queue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}
