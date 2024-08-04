package lol

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/davecgh/go-spew/spew"
)

const (
	Off = iota
	Fatal
	Error
	Warn
	Info
	Debug
	Trace
)

var LevelNames = []string{
	"off",
	"fatal",
	"error",
	"warn",
	"info",
	"debug",
	"trace",
}

type (
	// LevelPrinter defines a set of terminal printing primitives that output with
	// extra data, time, log logLevelList, and code location

	// Ln prints lists of interfaces with spaces in between
	Ln func(a ...interface{})
	// F prints like fmt.Println surrounded by log details
	F func(format string, a ...interface{})
	// S prints a spew.Sdump for an enveloper slice
	S func(a ...interface{})
	// C accepts a function so that the extra computation can be avoided if it is
	// not being viewed
	C func(closure func() string)
	// Chk is a shortcut for printing if there is an error, or returning true
	Chk func(e error) bool
	// Err is a pass-through function that uses fmt.Errorf to construct an error
	// and returns the error after printing it to the log
	Err          func(format string, a ...interface{}) error
	LevelPrinter struct {
		Ln
		F
		S
		C
		Chk
		Err
	}
	LevelSpec struct {
		ID        int
		Name      string
		Colorizer func(a ...interface{}) string
	}

	// Entry is a log entry to be printed as json to the log file
	Entry struct {
		Time         time.Time
		Level        string
		Package      string
		CodeLocation string
		Text         string
	}
)

var (
	// sep is just a convenient shortcut for this very longwinded expression
	sep = string(os.PathSeparator)
	// writer can be swapped out for any io.*writer* that you want to use instead of
	// stdout.
	writer io.Writer = os.Stderr
	// LevelSpecs specifies the id, string name and color-printing function
	LevelSpecs = []LevelSpec{
		{Off, "   ", fmt.Sprint},
		{Fatal, "FTL", fmt.Sprint},
		{Error, "ERR", fmt.Sprint},
		{Warn, "WRN", fmt.Sprint},
		{Info, "INF", fmt.Sprint},
		{Debug, "DBG", fmt.Sprint},
		{Trace, "TRC", fmt.Sprint},
	}
)

// Log is a set of log printers for the various Level items.
type Log struct {
	F, E, W, I, D, T LevelPrinter
}

type Check struct {
	F, E, W, I, D, T Chk
}
type Errorf struct {
	F, E, W, I, D, T Err
}

type Logger struct {
	*Log
	*Check
	*Errorf
}

var Main *Logger

func init() {
	Main = &Logger{}
	SetLoggers(Info)
}

func SetLoggers(level int) {
	Main.Log, Main.Check, Main.Errorf = New(level, os.Stderr)
}

func SetLogLevel(level string) {
	for i := range LevelNames {
		if level == LevelNames[i] {
			SetLoggers(i)
			return
		}
	}
}

func JoinStrings(a ...any) (s string) {
	for i := range a {
		s += fmt.Sprint(a[i])
		if i < len(a)-1 {
			s += " "
		}
	}
	return
}

func GetPrinter(l int32, writer io.Writer) LevelPrinter {
	return LevelPrinter{
		Ln: func(a ...interface{}) {
			fmt.Fprintf(writer,
				"%s %s %s %s\n",
				UnixTimeAsFloat(),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				JoinStrings(a...),
				GetLoc(2),
			)
		},
		F: func(format string, a ...interface{}) {
			fmt.Fprintf(writer,
				"%s %s %s %s\n",
				UnixTimeAsFloat(),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				fmt.Sprintf(format, a...),
				GetLoc(2),
			)
		},
		S: func(a ...interface{}) {
			fmt.Fprintf(writer,
				"%s %s %s %s\n",
				UnixTimeAsFloat(),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				spew.Sdump(a...),
				GetLoc(2),
			)
		},
		C: func(closure func() string) {
			fmt.Fprintf(writer,
				"%s %s %s %s\n",
				UnixTimeAsFloat(),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
				closure(),
				GetLoc(2),
			)
		},
		Chk: func(e error) bool {
			if e != nil {
				fmt.Fprintf(writer,
					"%s %s %s %s\n",
					UnixTimeAsFloat(),
					LevelSpecs[l].Colorizer(LevelSpecs[l].Name),
					e.Error(),
					GetLoc(2),
				)
				return true
			}
			return false
		},
		Err: func(format string, a ...interface{}) error {
			fmt.Fprintf(writer,
				"%s %s %s %s\n",
				UnixTimeAsFloat(),
				LevelSpecs[l].Colorizer(LevelSpecs[l].Name, " "),
				fmt.Sprintf(format, a...),
				GetLoc(2),
			)
			return fmt.Errorf(format, a...)
		},
	}
}

func GetNullPrinter() LevelPrinter {
	return LevelPrinter{
		Ln:  func(a ...interface{}) {},
		F:   func(format string, a ...interface{}) {},
		S:   func(a ...interface{}) {},
		C:   func(closure func() string) {},
		Chk: func(e error) bool { return e != nil },
		Err: func(format string, a ...interface{}) error { return fmt.Errorf(format, a...) },
	}
}

func New(level int, writer io.Writer) (l *Log, c *Check, errorf *Errorf) {
	l = &Log{
		F: GetNullPrinter(),
		E: GetNullPrinter(),
		W: GetNullPrinter(),
		I: GetNullPrinter(),
		D: GetNullPrinter(),
		T: GetNullPrinter(),
	}
	c = &Check{}
	errorf = &Errorf{}
	switch {
	case level >= Trace:
		l.T = GetPrinter(Trace, writer)
		fallthrough
	case level >= Debug:
		l.D = GetPrinter(Debug, writer)
		fallthrough
	case level >= Info:
		l.I = GetPrinter(Info, writer)
		fallthrough
	case level >= Warn:
		l.W = GetPrinter(Warn, writer)
		fallthrough
	case level >= Error:
		l.E = GetPrinter(Error, writer)
		fallthrough
	case level >= Fatal:
		l.F = GetPrinter(Fatal, writer)
	}
	c = &Check{
		F: l.F.Chk,
		E: l.E.Chk,
		W: l.W.Chk,
		I: l.I.Chk,
		D: l.D.Chk,
		T: l.T.Chk,
	}
	errorf = &Errorf{
		F: l.F.Err,
		E: l.E.Err,
		W: l.W.Err,
		I: l.I.Err,
		D: l.D.Err,
		T: l.T.Err,
	}
	return
}

// UnixTimeAsFloat e
func UnixTimeAsFloat() (s string) {
	timeText := fmt.Sprint(time.Now().UnixNano())
	lt := len(timeText)
	lb := lt + 1
	var timeBytes = make([]byte, lb)
	copy(timeBytes[lb-9:lb], timeText[lt-9:lt])
	timeBytes[lb-10] = '.'
	lb -= 10
	lt -= 9
	copy(timeBytes[:lb], timeText[:lt])
	return fmt.Sprint(string(timeBytes))
}

func GetLoc(skip int) (output string) {
	_, file, line, _ := runtime.Caller(skip)
	output = fmt.Sprint(
		file, ":", line,
	)
	return
}
