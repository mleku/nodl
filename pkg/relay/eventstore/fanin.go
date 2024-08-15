package eventstore

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/util/context"
)

// FanIn listens to a provided group of channels and forwards anything from them to a returned
// "out" channel.
func FanIn(c context.T, in ...event.C) (out event.C) {
	out = make(event.C)
	go func() {
		for {
			select {
			case <-c.Done():
				// close(out)
				for i := range in {
					// ensure the channel is emptied first, let the senders terminate.
					select {
					case v := <-in[i]:
						if v != nil {
							log.T.S("draining results channel")
							// drain channel
							for range in[i] {
							}
							close(in[i])
						}
					default:
						close(in[i])
					}
				}
			}
		}
	}()
	for i := range in {
		go func() {
			for {
				select {
				case v := <-in[i]:
					log.T.S("forwarding result to main result channel")
					out <- v
				case <-c.Done():
					return
				}
			}
		}()
	}
	return
}
