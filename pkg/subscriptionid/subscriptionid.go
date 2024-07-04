package subscriptionid

import (
	"errors"

	"github.com/mleku/nodl/pkg/text"
)

type T B

// IsValid returns true if the subscription id is between 1 and 64 characters.
// Invalid means too long or not present.
func (si T) IsValid() bool { return len(si) <= 64 && len(si) > 0 }

// New inspects a string and converts to T if it is
// valid. Invalid means length == 0 or length > 64.
func New(s string) (T, error) {
	si := T(s)
	if si.IsValid() {
		return si, nil
	} else {
		// remove invalid return value
		return si[:0], errors.New(
			"invalid subscription ID - either < 0 or > 64 char length")
	}
}

func (si T) Marshal(dst B) (b B, err error) {
	b = dst
	b = append(b, '"')
	b = text.NostrEscape(b, si)
	b = append(b, '"')
	return
}

func (si T) MarshalJSON(dst B) (b B, err error) {
	ue := text.NostrEscape(nil, si)
	if len(ue) < 1 || len(ue) > 64 {
		err = errorf.E("invalid subscription ID, must be between 1 and 64 "+
			"characters, got %d (possibly due to escaping)", len(ue))
		return
	}
	b = dst
	b = append(b, '"')
	b = append(b, ue...)
	b = append(b, '"')
	return
}

func (si T) UnmarshalJSON(b B) (ta any, rem B, err error) {
	var openQuotes, escaping bool
	var start int
	rem = b
	for i := range rem {
		if !openQuotes && rem[i] == '"' {
			openQuotes = true
			start = i + 1
		} else if openQuotes {
			if !escaping && rem[i] == '\\' {
				escaping = true
			} else if rem[i] == '"' {
				if !escaping {
					si = text.NostrUnescape(rem[start:i])
					ta = si
					rem = rem[i+1:]
					return
				} else {
					escaping = false
				}
			} else {
				escaping = false
			}
		}
	}
	return
}
