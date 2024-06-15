package lol_test

import (
	"errors"
	"testing"

	"github.com/mleku/nodl/pkg/lol"
)

func TestGetLogger(t *testing.T) {
	for i := 0; i < 100; i++ {
		lol.SetLogLevel(lol.Trace)
		log.T.Ln("testing log level", lol.LevelSpecs[lol.Trace].Name)
		log.D.Ln("testing log level", lol.LevelSpecs[lol.Debug].Name)
		log.I.Ln("testing log level", lol.LevelSpecs[lol.Info].Name)
		log.W.Ln("testing log level", lol.LevelSpecs[lol.Warn].Name)
		log.E.F("testing log level %s", lol.LevelSpecs[lol.Error].Name)
		log.F.Ln("testing log level", lol.LevelSpecs[lol.Fatal].Name)
		chk.F(errors.New("dummy error as fatal"))
		chk.E(errors.New("dummy error as error"))
		chk.W(errors.New("dummy error as warning"))
		chk.I(errors.New("dummy error as info"))
		chk.D(errors.New("dummy error as debug"))
		chk.T(errors.New("dummy error as trace"))
		log.I.Ln("log.I.Err",
			log.I.Err("format string %d '%s'", 5, "testing") != nil)
		log.I.Chk(errors.New("dummy information check"))
		log.I.Chk(nil)
		log.I.S("`backtick wrapped string`", t)
	}
}
