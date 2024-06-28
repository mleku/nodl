package sentinel

import (
	"github.com/mleku/nodl/pkg/envelopes"
)

func Read(label string, b B) (env envelopes.I, err error) {
	// rem := b
	// switch label {
	// case authenvelope.L:
	// 	if rem[0] == '{' {
	// 		// response
	// 		if env, rem, err = authenvelope.UnmarshalResponse(rem); chk.E(err) {
	// 			return
	// 		}
	// 	} else if rem[0] == '"' {
	// 		// challenge
	// 		if env, rem, err = authenvelope.UnmarshalChallenge(rem); chk.E(err) {
	// 			return
	// 		}
	// 	}
	// 	err = fmt.Errorf("invalid: envelope label %s neither challenge or response\n%s",
	// 		label, b)
	// 	return
	// }
	return
}
