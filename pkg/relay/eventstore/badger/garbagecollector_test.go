package badger

// import (
// 	"os"
// 	"sync"
// 	"testing"
// 	"time"

// 	"git.replicatr.dev/pkg/relay/bech32encoding"
// 	"git.replicatr.dev/pkg/relay/context"
// 	"git.replicatr.dev/pkg/relay/event"
// 	"git.replicatr.dev/pkg/relay/eventid"
// 	"git.replicatr.dev/pkg/relay/filter"
// 	"git.replicatr.dev/pkg/relay/keys"
// 	"git.replicatr.dev/pkg/relay/tag"
// 	"git.replicatr.dev/pkg/relay/tests"
// 	"lukechampine.com/frand"
// 	"git.replicatr.dev/pkg/util/interrupt"
// 	"git.replicatr.dev/pkg/util/qu"
// )

// type Counter struct {
// 	id        *eventid.T
// 	requested int
// }

// func TestGarbageCollector(t *testing.T) {
// 	var (
// 		err            error
// 		sec            string
// 		mx             sync.Mutex
// 		counter        []Counter
// 		total          int
// 		MaxContentSize = 4096
// 		TotalSize      = 1000000
// 		MaxDelay       = time.Second / 5
// 	)
// 	sec = keys.GeneratePrivateKey()
// 	var nsec string

// 	if nsec, err = bech32encoding.HexToNsec(sec); err != nil {
// 		panic(err)
// 	}
// 	go be.GarbageCollector()
// 	interrupt.AddHandler(func() {
// 		cancel()
// 		chk.E(os.RemoveAll(path))
// 	})
// end:
// 	for {
// 		select {
// 		case <-c.Done():
// 			log.I.Ln("context canceled")
// 			return
// 		default:
// 		}
// 		mx.Lock()
// 		if total > TotalSize {
// 			mx.Unlock()
// 			cancel()
// 			return
// 		}
// 		mx.Unlock()
// 		newEvent := qu.T()
// 		go func() {
// 			ticker := time.NewTicker(time.Second)
// 			var fetchIDs []*eventid.T
// 			// start fetching loop
// 			for {
// 				select {
// 				case <-newEvent:
// 					// make new request, not necessarily from existing... bias rng
// 					// factor by request count
// 					mx.Lock()
// 					for i := range counter {
// 						rn := frand.Intn(256)
// 						// multiply this number by the number of accesses the event
// 						// has and request every event that gets over 50% so that we
// 						// create a bias towards already requested.
// 						if counter[i].requested+rn > 192 {
// 							log.T.Ln("counter", counter[i].requested, "+", rn, "=",
// 								counter[i].requested+rn)
// 							// log.T.Ln("adding to fetchIDs")
// 							counter[i].requested++
// 							fetchIDs = append(fetchIDs, counter[i].id)
// 						}
// 					}
// 					if len(fetchIDs) > 0 {
// 						log.T.Ln("fetchIDs", len(fetchIDs), fetchIDs)
// 					}
// 					mx.Unlock()
// 				case <-ticker.C:
// 					// copy out current list of events to request
// 					mx.Lock()
// 					log.T.Ln("ticker", len(fetchIDs))
// 					ids := make(tag.T, len(fetchIDs))
// 					for i := range fetchIDs {
// 						ids[i] = fetchIDs[i].String()
// 					}
// 					fetchIDs = fetchIDs[:0]
// 					mx.Unlock()
// 					if len(ids) > 0 {
// 						for i := range ids {
// 							go func(i int) {
// 								sc, scancel := context.Cancel(context.Bg())
// 								var ch event.C
// 								ch, err = be.QueryEvents(sc,
// 									&filter.T{IDs: tag.T{ids[i]}})
// 								go func() {
// 									// receive the results
// 									select {
// 									case <-time.After(time.Second):
// 										log.I.Ln("cancel")
// 										scancel()
// 									case <-ch:
// 										log.T.Ln("received event")
// 									case <-sc.Done():
// 										log.I.Ln("subscription done")
// 									case <-c.Done():
// 										log.T.Ln("context canceled")
// 										return
// 									}
// 								}()
// 							}(i)
// 						}
// 					}
// 				case <-c.Done():
// 					log.I.Ln("context canceled")
// 					return
// 				}
// 			}
// 		}()
// 		var ev *event.T
// 		var bs int
// 	out:
// 		for {
// 			select {
// 			case <-c.Done():
// 				return
// 			default:
// 			}
// 			if ev, bs, err = tests.GenerateEvent(sec, MaxContentSize); chk.E(err) {
// 				return
// 			}
// 			mx.Lock()
// 			counter = append(counter, Counter{id: &ev.ID, requested: 1})
// 			total += bs
// 			if total > TotalSize {
// 				mx.Unlock()
// 				cancel()
// 				break out
// 			}
// 			mx.Unlock()
// 			newEvent.Signal()
// 			sc, _ := context.Timeout(c, 2*time.Second)
// 			if err = be.SaveEvent(sc, ev); chk.E(err) {
// 				continue end
// 			}
// 			delay := frand.Intn(int(MaxDelay))
// 			log.T.Ln("waiting between", delay, "ns")
// 			if delay == 0 {
// 				continue
// 			}
// 			select {
// 			case <-c.Done():
// 				return
// 			case <-time.After(time.Duration(delay)):
// 			}
// 		}
// 		select {
// 		case <-c.Done():
// 		}
// 	}
// }