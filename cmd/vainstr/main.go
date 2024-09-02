package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	. "nostr.mleku.dev"

	"ec.mleku.dev/v2/bech32"
	"github.com/alexflint/go-arg"
	"nostr.mleku.dev/codec/bech32encoding"
	"nostr.mleku.dev/crypto/p256k"
	"util.mleku.dev/atomic"
	"util.mleku.dev/interrupt"
	"util.mleku.dev/qu"
)

type Result struct {
	sec, pub B
	npub     B
}

var prefix = append(bech32encoding.PubHRP, '1')

const (
	PositionBeginning = iota
	PositionContains
	PositionEnding
)

var args struct {
	String   string `arg:"positional" help:"the string you want to appear in the npub"`
	Position string `arg:"positional" default:"end" help:"[begin|contain|end] default: end"`
	Threads  int    `help:"number of threads to mine with - defaults to using all CPU threads available"`
}

func main() {
	arg.MustParse(&args)
	if args.String == "" {
		_, _ = fmt.Fprintln(os.Stderr,
			`Usage: vainstr [--threads THREADS] [STRING [POSITION]]

Positional arguments:
  STRING                 the string you want to appear in the npub
  POSITION               [begin|contain|end] default: end

Options:
  --threads THREADS      number of threads to mine with - defaults to using all CPU threads available
  --help, -h             display this help and exit`)
		os.Exit(0)
	}
	var where int
	canonical := strings.ToLower(args.Position)
	switch {
	case strings.HasPrefix(canonical, "begin"):
		where = PositionBeginning
	case strings.Contains(canonical, "contain"):
		where = PositionContains
	case strings.HasSuffix(canonical, "end"):
		where = PositionEnding
	}
	if args.Threads == 0 {
		args.Threads = runtime.NumCPU()
	}
	if err := Vanity(B(args.String), where, args.Threads); Chk.F(err) {
	}
}

func Vanity(str B, where int, threads int) (e error) {
	// check the string has valid bech32 ciphers
	for i := range str {
		wrong := true
		for j := range bech32.Charset {
			if str[i] == bech32.Charset[j] {
				wrong = false
				break
			}
		}
		if wrong {
			return fmt.Errorf("found invalid character '%c' only ones from '%s' allowed\n",
				str[i], bech32.Charset)
		}
	}
	started := time.Now()
	quit, shutdown := qu.T(), qu.T()
	resC := make(chan Result)
	interrupt.AddHandler(func() {
		// this will stop work if CTRL-C or Interrupt signal from OS.
		shutdown.Q()
	})
	var wg sync.WaitGroup
	counter := atomic.NewInt64(0)
	for i := 0; i < threads; i++ {
		go mine(str, where, quit, resC, &wg, counter)
	}
	tick := time.NewTicker(time.Second * 5)
	var res Result
out:
	for {
		select {
		case <-tick.C:
			workingFor := time.Now().Sub(started)
			wm := workingFor % time.Second
			workingFor -= wm
			Log.I.F("working for %v, attempts %d",
				workingFor, counter.Load())
		case r := <-resC:
			// one of the workers found the solution
			res = r
			// tell the others to stop
			quit.Q()
			break out
		case <-shutdown.Wait():
			quit.Q()
			os.Exit(0)
		}
	}

	// wait for all of the workers to stop
	wg.Wait()

	fmt.Printf("# generated in %d attempts using %d threads, taking %v\n",
		counter.Load(), args.Threads, time.Now().Sub(started))
	nsec, _ := bech32encoding.BinToNsec(res.sec)
	fmt.Printf("NSEC = %s\nNPUB = %s\n", nsec, res.npub)
	fmt.Printf("SEC = %0x\nPUB = %0x\n", res.sec, res.pub)
	return
}

func mine(str B, where int, quit qu.C, resC chan Result, wg *sync.WaitGroup,
	counter *atomic.Int64) {

	wg.Add(1)
	var r Result
	var err error
	found := false
	atstart := append(prefix, str...)
	s := p256k.NewKeygen()
	var pkb B
out:
	for {
		select {
		case <-quit:
			wg.Done()
			if found {
				if pkb[0] == 3 {
					s.Negate()
				}
				sk, pk := s.KeyPairBytes()
				r.sec = sk
				r.pub = pk
				// send back the result
				resC <- r
			}
			break out
		default:
		}
		counter.Inc()
		pkb, err = s.Generate()
		if r.npub, err = bech32encoding.BinToNpub(pkb[1:]); Chk.E(err) {
			Log.F.Ln("fatal error generating npub: %s\n", err)
			break out
		}
		switch where {
		case PositionBeginning:
			if bytes.HasPrefix(r.npub, atstart) {
				found = true
				quit.Q()
			}
		case PositionEnding:
			if bytes.HasSuffix(r.npub, str) {
				found = true
				quit.Q()
			}
		case PositionContains:
			if bytes.Contains(r.npub, str) {
				found = true
				quit.Q()
			}
		}
	}
}
