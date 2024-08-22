package badgerbadger

import (
	"nostr.mleku.dev/codec/eventid"
)

type Counter struct {
	id        *eventid.T
	size      int
	requested int
}

// func TestBackend(t *testing.T) {
// 	var (
// 		err            E
// 		sec            B
// 		mx             sync.Mutex
// 		counter        []Counter
// 		total          int
// 		MaxContentSize = 16384
// 		TotalSize      = 10000000
// 		MaxDelay       = time.Second / 4
// 		HW             = 95
// 		LW             = 90
// 		// fill rate capped to size of differerce between high and low water mark
// 		diff = TotalSize / 100 * (HW - LW) / 100 / 100
// 	)
// 	// todo: this isn't working currently
// 	sec = keys.GenerateSecretKeyHex()
// 	var nsec B
// 	if nsec, err = bech32encoding.HexToNsec(sec); chk.E(err) {
// 		panic(err)
// 	}
// 	log.T.Ln("signing with", nsec)
// 	c, cancel := context.Cancel(context.Bg())
// 	var wg sync.WaitGroup
// 	defer cancel()
// 	// create L1 with cache management settings enabled
// 	path1 := filepath.Join(os.TempDir(), fmt.Sprintf("%0x", frand.Bytes(8)))
// 	b1 := badger.GetBackend(c, &wg, true, app.MaxMessageSize,
// 		TotalSize/100, LW, HW, 2)
// 	// create L2 with no cache management
// 	path2 := filepath.Join(os.TempDir(), fmt.Sprintf("%0x", frand.Bytes(8)))
// 	b2 := badger.GetBackend(c, &wg, false, app.MaxMessageSize, 0)
// 	// Respond to interrupt signal and clean up after interrupt or end of test.
// 	interrupt.AddHandler(func() {
// 		cancel()
// 		chk.E(os.RemoveAll(path1))
// 		chk.E(os.RemoveAll(path2))
// 	})
// 	// now join them together in a 2 level eventstore
// 	twoLevel := l2.Backend{
// 		Ctx: c,
// 		WG:  &wg,
// 		L1:  b1,
// 		L2:  b2,
// 	}
// 	if err = twoLevel.Init(""); chk.E(err) {
// 		t.Fatal()
// 	}
// 	// start GC
// 	go b1.GarbageCollector()
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
// 					var sum int
// 					for i := range counter {
// 						rn := frand.Intn(256)
// 						if sum > diff {
// 							// don't overfill
// 							break
// 						}
// 						// multiply this number by the number of accesses the event
// 						// has and request every event that gets over 50% so that we
// 						// create a bias towards already requested.
// 						if counter[i].requested+rn > 192 {
// 							log.T.Ln("counter", counter[i].requested, "+", rn,
// 								"=",
// 								counter[i].requested+rn)
// 							// log.T.Ln("adding to fetchIDs")
// 							counter[i].requested++
// 							fetchIDs = append(fetchIDs, counter[i].id)
// 							sum += counter[i].size
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
// 					ids := tag.New("")
// 					for i := range fetchIDs {
// 						ids.Field[i] = fetchIDs[i].Bytes()
// 					}
// 					fetchIDs = fetchIDs[:0]
// 					mx.Unlock()
// 					if ids.Len() > 0 {
// 						for i := range ids.Field {
// 							// go func(i int) {
// 							sc, scancel := context.Cancel(context.Bg())
// 							_ = scancel
// 							var ch []*event.T
// 							f := B(ids.Field[i])
// 							ch, err = twoLevel.QueryEvents(sc, &filter.T{IDs: tag.New(f)})
// 							_ = ch
// 							// go func() {
// 							// 	// receive the results
// 							// 	select {
// 							// 	case <-time.After(time.Second):
// 							// 		// log.I.Ln("cancel")
// 							// 		scancel()
// 							// 	// case <-ch:
// 							// 	// 	log.T.Ln("received event")
// 							// 	case <-sc.Done():
// 							// 		log.I.Ln("subscription done")
// 							// 	case <-c.Done():
// 							// 		log.T.Ln("context canceled")
// 							// 		return
// 							// 	}
// 							// }()
// 							// }(i)
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
// 			counter = append(counter,
// 				Counter{id: eventid.NewWith(ev.ID), size: bs, requested: 1})
// 			total += bs
// 			if total > TotalSize {
// 				mx.Unlock()
// 				cancel()
// 				break out
// 			}
// 			mx.Unlock()
// 			newEvent.Signal()
// 			sc, _ := context.Timeout(c, 2*time.Second)
// 			if err = twoLevel.SaveEvent(sc, ev); chk.E(err) {
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
