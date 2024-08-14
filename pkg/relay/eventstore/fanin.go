package eventstore

import (
	"git.replicatr.dev/pkg/codec/event"
	"git.replicatr.dev/pkg/util/context"
)

// FanIn listens to a provided group of channels and forwards anything from them to a returned
// "out" channel.
//
// Cancel the context to close all in and the out channel.
//
// Will automatically close the out channel if all in channels are closed.
func FanIn(c context.T, in ...event.C) (out event.C) {
	go func() {
		for {
			select {
			case <-c.Done():
				close(out)
				for i := range in {
					// ensure the channel is emptied first, let the senders terminate.
					select {
					case v := <-in[i]:
						if v != nil {
							// drain channel
							for range in[i] {
							}
							close(in[i])
						}
					default:
					}
					close(in[i])
				}
			}
		}
	}()
	for i := range in {
		go func() {
			for {
				select {
				case v := <-in[i]:
					out <- v
				case <-c.Done():
					close(out)
				}
			}
		}()
	}
	return
}
