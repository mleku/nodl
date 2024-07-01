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
// valid. Invalid means length < 0 and <= 64 (hex encoded 256 bit hash).
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
