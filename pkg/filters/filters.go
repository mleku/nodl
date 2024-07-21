package filters

import (
	"github.com/mleku/nodl/pkg/event"
	"github.com/mleku/nodl/pkg/filter"
)

type T struct {
	F []*filter.T
}

func (f *T) Len() int { return len(f.F) }

func New() (f *T) { return &T{} }

func (f *T) Match(event *event.T) bool {
	for _, f := range f.F {
		if f.Matches(event) {
			return true
		}
	}
	return false
}

func (f *T) MarshalJSON(dst B) (b B, err error) {
	b = dst
	b = append(b, '[')
	end := len(f.F) - 1
	for i := range f.F {
		if b, err = f.F[i].MarshalJSON(b); chk.E(err) {
			return
		}
		if i < end {
			b = append(b, ',')
		}
	}
	b = append(b, ']')
	return
}

func (f *T) UnmarshalJSON(b B) (r B, err error) {
	r = b[:]
	for len(r) > 0 {
		switch r[0] {
		case '[':
			if len(r) > 1 && r[1] == ']' {
				r = r[1:]
				return
			}
			r = r[1:]
			ffa := filter.New()
			if r, err = ffa.UnmarshalJSON(r); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
			// continue
		case ',':
			r = r[1:]
			if len(r) > 1 && r[1] == ']' {
				r = r[1:]
				return
			}
			ffa := filter.New()
			if r, err = ffa.UnmarshalJSON(r); chk.E(err) {
				return
			}
			f.F = append(f.F, ffa)
		// next
		case ']':
			r = r[1:]
			// the end
			return
		}
	}
	return
}

func GenFilters(n int) (ff *T, err error) {
	ff = &T{}
	for _ = range n {
		var f *filter.T
		if f, err = filter.GenFilter(); chk.E(err) {
			return
		}
		ff.F = append(ff.F, f)
	}
	return
}
