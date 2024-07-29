package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"ec.mleku.dev/v2/bech32"
	"git.replicatr.dev/pkg/codec/bech32encoding"
	"git.replicatr.dev/pkg/crypto/p256k"
	"git.replicatr.dev/pkg/util/atomic"
	"git.replicatr.dev/pkg/util/hex"
	"git.replicatr.dev/pkg/util/interrupt"
	"git.replicatr.dev/pkg/util/qu"
	"github.com/alexflint/go-arg"
)

var prefix = append(bech32encoding.PubHRP, '1')

const (
	PositionBeginning = iota
	PositionContains
	PositionEnding
)

type Result struct {
	sec, pub B
	npub     B
}

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
	if err := Vanity(B(args.String), where, args.Threads); err != nil {
		log.F.F("error: %s", err)
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
		log.D.F("starting up worker %d", i)
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
			fmt.Printf("working for %v, attempts %d\n",
				workingFor, counter.Load())
		case r := <-resC:
			// one of the workers found the solution
			res = r
			// tell the others to stop
			quit.Q()
			break out
		case <-shutdown.Wait():
			quit.Q()
			log.I.Ln("\rinterrupt signal received")
			os.Exit(0)
		}
	}

	// wait for all of the workers to stop
	wg.Wait()

	fmt.Printf("generated in %d attempts using %d threads, taking %v\n",
		counter.Load(), args.Threads, time.Now().Sub(started))
	log.D.Ln(
		"generated key pair:\n"+
			"\nhex:\n"+
			"\tsecret: %s\n"+
			"\tpublic: %s\n\n",
		hex.Enc(res.sec),
		hex.Enc(res.pub),
	)
	nsec, _ := bech32encoding.BinToNsec(res.sec)
	fmt.Printf("\nNSEC = %s\nNPUB = %s\n\n",
		nsec, res.npub)
	return
}

func mine(str B, where int, quit qu.C, resC chan Result, wg *sync.WaitGroup, counter *atomic.Int64) {

	wg.Add(1)
	var r Result
	var err error
	found := false
	atstart := append(prefix, str...)
out:
	for {
		select {
		case <-quit:
			wg.Done()
			if found {
				// send back the result
				log.D.Ln("sending back result\n")
				resC <- r
				log.D.Ln("sent\n")
			} else {
				log.D.Ln("other thread found it\n")
			}
			break out
		default:
		}
		counter.Inc()
		if r.sec, r.pub, err = p256k.GenSecBytes(); chk.E(err) {
			log.E.Ln("error generating key: '%v' worker stopping", err)
			break out
		}
		r.npub, err = bech32encoding.BinToNpub(r.pub)
		if err != nil {
			log.E.Ln("fatal error generating npub: %s\n", err)
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
