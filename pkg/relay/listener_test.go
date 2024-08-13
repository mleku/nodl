package relay

import (
	"math/rand"
	"testing"

	"git.replicatr.dev/pkg/codec/filter"
	"git.replicatr.dev/pkg/codec/kind"
	"git.replicatr.dev/pkg/codec/kinds"
	"github.com/stretchr/testify/require"
)

func TestListenerSetupAndRemoveOnce(t *testing.T) {
	rl := New()

	ws1 := &WebSocket{}
	ws2 := &WebSocket{}

	f1 := filter.T{Kinds: kinds.New(kind.New(1))}
	f2 := filter.T{Kinds: kinds.New(kind.New(2))}
	f3 := filter.T{Kinds: kinds.New(kind.New(3))}

	rl.clients[ws1] = nil
	rl.clients[ws2] = nil

	var cancel func(cause error) = nil

	t.Run("adding listeners", func(t *testing.T) {
		rl.addListener(ws1, "1a", rl, f1, cancel)
		rl.addListener(ws1, "1b", rl, f2, cancel)
		rl.addListener(ws2, "2a", rl, f3, cancel)
		rl.addListener(ws1, "1c", rl, f3, cancel)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"1a", cancel, 0, rl},
				{"1b", cancel, 1, rl},
				{"1c", cancel, 3, rl},
			},
			ws2: {
				{"2a", cancel, 2, rl},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"1a", f1, ws1},
			{"1b", f2, ws1},
			{"2a", f3, ws2},
			{"1c", f3, ws1},
		}, rl.listeners)
	})

	t.Run("removing a client", func(t *testing.T) {
		rl.removeClientAndListeners(ws1)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws2: {
				{"2a", cancel, 0, rl},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"2a", f3, ws2},
		}, rl.listeners)
	})
}

func TestListenerMoreConvolutedCase(t *testing.T) {
	rl := New()

	ws1 := &WebSocket{}
	ws2 := &WebSocket{}
	ws3 := &WebSocket{}
	ws4 := &WebSocket{}

	f1 := filter.T{Kinds: kinds.New(kind.New(1))}
	f2 := filter.T{Kinds: kinds.New(kind.New(2))}
	f3 := filter.T{Kinds: kinds.New(kind.New(3))}

	rl.clients[ws1] = nil
	rl.clients[ws2] = nil
	rl.clients[ws3] = nil
	rl.clients[ws4] = nil

	var cancel func(cause error) = nil

	t.Run("adding listeners", func(t *testing.T) {
		rl.addListener(ws1, "c", rl, f1, cancel)
		rl.addListener(ws2, "b", rl, f2, cancel)
		rl.addListener(ws3, "a", rl, f3, cancel)
		rl.addListener(ws4, "d", rl, f3, cancel)
		rl.addListener(ws2, "b", rl, f1, cancel)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rl},
			},
			ws2: {
				{"b", cancel, 1, rl},
				{"b", cancel, 4, rl},
			},
			ws3: {
				{"a", cancel, 2, rl},
			},
			ws4: {
				{"d", cancel, 3, rl},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"b", f2, ws2},
			{"a", f3, ws3},
			{"d", f3, ws4},
			{"b", f1, ws2},
		}, rl.listeners)
	})

	t.Run("removing a client", func(t *testing.T) {
		rl.removeClientAndListeners(ws2)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rl},
			},
			ws3: {
				{"a", cancel, 2, rl},
			},
			ws4: {
				{"d", cancel, 1, rl},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"d", f3, ws4},
			{"a", f3, ws3},
		}, rl.listeners)
	})

	t.Run("reorganize the first case differently and then remove again", func(t *testing.T) {
		rl.clients = map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 1, rl},
			},
			ws2: {
				{"b", cancel, 2, rl},
				{"b", cancel, 4, rl},
			},
			ws3: {
				{"a", cancel, 0, rl},
			},
			ws4: {
				{"d", cancel, 3, rl},
			},
		}
		rl.listeners = []listener{
			{"a", f3, ws3},
			{"c", f1, ws1},
			{"b", f2, ws2},
			{"d", f3, ws4},
			{"b", f1, ws2},
		}

		rl.removeClientAndListeners(ws2)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 1, rl},
			},
			ws3: {
				{"a", cancel, 0, rl},
			},
			ws4: {
				{"d", cancel, 2, rl},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"a", f3, ws3},
			{"c", f1, ws1},
			{"d", f3, ws4},
		}, rl.listeners)
	})
}

func TestListenerMoreStuffWithMultipleRelays(t *testing.T) {
	rl := New()

	ws1 := &WebSocket{}
	ws2 := &WebSocket{}
	ws3 := &WebSocket{}
	ws4 := &WebSocket{}

	f1 := filter.T{Kinds: kinds.New(kind.New(1))}
	f2 := filter.T{Kinds: kinds.New(kind.New(2))}
	f3 := filter.T{Kinds: kinds.New(kind.New(3))}

	rlx := New()
	rly := New()
	rlz := New()

	rl.clients[ws1] = nil
	rl.clients[ws2] = nil
	rl.clients[ws3] = nil
	rl.clients[ws4] = nil

	var cancel func(cause error) = nil

	t.Run("adding listeners", func(t *testing.T) {
		rl.addListener(ws1, "c", rlx, f1, cancel)
		rl.addListener(ws2, "b", rly, f2, cancel)
		rl.addListener(ws3, "a", rlz, f3, cancel)
		rl.addListener(ws4, "d", rlx, f3, cancel)
		rl.addListener(ws4, "e", rlx, f3, cancel)
		rl.addListener(ws3, "a", rlx, f3, cancel)
		rl.addListener(ws4, "e", rly, f3, cancel)
		rl.addListener(ws3, "f", rly, f3, cancel)
		rl.addListener(ws1, "g", rlz, f1, cancel)
		rl.addListener(ws2, "g", rlz, f2, cancel)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rlx},
				{"g", cancel, 1, rlz},
			},
			ws2: {
				{"b", cancel, 0, rly},
				{"g", cancel, 2, rlz},
			},
			ws3: {
				{"a", cancel, 0, rlz},
				{"a", cancel, 3, rlx},
				{"f", cancel, 2, rly},
			},
			ws4: {
				{"d", cancel, 1, rlx},
				{"e", cancel, 2, rlx},
				{"e", cancel, 1, rly},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"d", f3, ws4},
			{"e", f3, ws4},
			{"a", f3, ws3},
		}, rlx.listeners)

		require.Equal(t, []listener{
			{"b", f2, ws2},
			{"e", f3, ws4},
			{"f", f3, ws3},
		}, rly.listeners)

		require.Equal(t, []listener{
			{"a", f3, ws3},
			{"g", f1, ws1},
			{"g", f2, ws2},
		}, rlz.listeners)
	})

	t.Run("removing a subscription id", func(t *testing.T) {
		// removing 'd' from ws4
		rl.clients[ws4][0].cancel = func(cause error) {} // set since removing will call it
		rl.removeListenerId(ws4, "d")

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rlx},
				{"g", cancel, 1, rlz},
			},
			ws2: {
				{"b", cancel, 0, rly},
				{"g", cancel, 2, rlz},
			},
			ws3: {
				{"a", cancel, 0, rlz},
				{"a", cancel, 1, rlx},
				{"f", cancel, 2, rly},
			},
			ws4: {
				{"e", cancel, 1, rly},
				{"e", cancel, 2, rlx},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"a", f3, ws3},
			{"e", f3, ws4},
		}, rlx.listeners)

		require.Equal(t, []listener{
			{"b", f2, ws2},
			{"e", f3, ws4},
			{"f", f3, ws3},
		}, rly.listeners)

		require.Equal(t, []listener{
			{"a", f3, ws3},
			{"g", f1, ws1},
			{"g", f2, ws2},
		}, rlz.listeners)
	})

	t.Run("removing another subscription id", func(t *testing.T) {
		// removing 'a' from ws3
		rl.clients[ws3][0].cancel = func(cause error) {} // set since removing will call it
		rl.clients[ws3][1].cancel = func(cause error) {} // set since removing will call it
		rl.removeListenerId(ws3, "a")

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rlx},
				{"g", cancel, 1, rlz},
			},
			ws2: {
				{"b", cancel, 0, rly},
				{"g", cancel, 0, rlz},
			},
			ws3: {
				{"f", cancel, 2, rly},
			},
			ws4: {
				{"e", cancel, 1, rly},
				{"e", cancel, 1, rlx},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"e", f3, ws4},
		}, rlx.listeners)

		require.Equal(t, []listener{
			{"b", f2, ws2},
			{"e", f3, ws4},
			{"f", f3, ws3},
		}, rly.listeners)

		require.Equal(t, []listener{
			{"g", f2, ws2},
			{"g", f1, ws1},
		}, rlz.listeners)
	})

	t.Run("removing a connection", func(t *testing.T) {
		rl.removeClientAndListeners(ws2)

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rlx},
				{"g", cancel, 0, rlz},
			},
			ws3: {
				{"f", cancel, 0, rly},
			},
			ws4: {
				{"e", cancel, 1, rly},
				{"e", cancel, 1, rlx},
			},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
			{"e", f3, ws4},
		}, rlx.listeners)

		require.Equal(t, []listener{
			{"f", f3, ws3},
			{"e", f3, ws4},
		}, rly.listeners)

		require.Equal(t, []listener{
			{"g", f1, ws1},
		}, rlz.listeners)
	})

	t.Run("removing another subscription id", func(t *testing.T) {
		// removing 'e' from ws4
		rl.clients[ws4][0].cancel = func(cause error) {} // set since removing will call it
		rl.clients[ws4][1].cancel = func(cause error) {} // set since removing will call it
		rl.removeListenerId(ws4, "e")

		require.Equal(t, map[*WebSocket][]listenerSpec{
			ws1: {
				{"c", cancel, 0, rlx},
				{"g", cancel, 0, rlz},
			},
			ws3: {
				{"f", cancel, 0, rly},
			},
			ws4: {},
		}, rl.clients)

		require.Equal(t, []listener{
			{"c", f1, ws1},
		}, rlx.listeners)

		require.Equal(t, []listener{
			{"f", f3, ws3},
		}, rly.listeners)

		require.Equal(t, []listener{
			{"g", f1, ws1},
		}, rlz.listeners)
	})
}

func TestRandomListenerClientRemoving(t *testing.T) {
	rl := New()

	f := filter.T{Kinds: kinds.New(kind.New(1))}
	cancel := func(cause error) {}

	websockets := make([]*WebSocket, 0, 20)

	l := 0

	for i := 0; i < 20; i++ {
		ws := &WebSocket{}
		websockets = append(websockets, ws)
		rl.clients[ws] = nil
	}

	for j := 0; j < 20; j++ {
		for i := 0; i < 20; i++ {
			ws := websockets[i]
			w := string(rune(i + 65))

			if rand.Intn(2) < 1 {
				l++
				rl.addListener(ws, w+":"+string(rune(j+97)), rl, f, cancel)
			}
		}
	}

	require.Len(t, rl.clients, 20)
	require.Len(t, rl.listeners, l)

	for ws := range rl.clients {
		rl.removeClientAndListeners(ws)
	}

	require.Len(t, rl.clients, 0)
	require.Len(t, rl.listeners, 0)
}

func TestRandomListenerIdRemoving(t *testing.T) {
	rl := New()

	f := filter.T{Kinds: kinds.New(kind.New(1))}
	cancel := func(cause error) {}

	websockets := make([]*WebSocket, 0, 20)

	type wsid struct {
		ws *WebSocket
		id string
	}

	subs := make([]wsid, 0, 20*20)
	extra := 0

	for i := 0; i < 20; i++ {
		ws := &WebSocket{}
		websockets = append(websockets, ws)
		rl.clients[ws] = nil
	}

	for j := 0; j < 20; j++ {
		for i := 0; i < 20; i++ {
			ws := websockets[i]
			w := string(rune(i + 65))

			if rand.Intn(2) < 1 {
				id := w + ":" + string(rune(j+97))
				rl.addListener(ws, id, rl, f, cancel)
				subs = append(subs, wsid{ws, id})

				if rand.Intn(5) < 1 {
					rl.addListener(ws, id, rl, f, cancel)
					extra++
				}
			}
		}
	}

	require.Len(t, rl.clients, 20)
	require.Len(t, rl.listeners, len(subs)+extra)

	rand.Shuffle(len(subs), func(i, j int) {
		subs[i], subs[j] = subs[j], subs[i]
	})
	for _, wsidToRemove := range subs {
		rl.removeListenerId(wsidToRemove.ws, wsidToRemove.id)
	}

	require.Len(t, rl.listeners, 0)
	require.Len(t, rl.clients, 20)
	for _, specs := range rl.clients {
		require.Len(t, specs, 0)
	}
}
