package kinds

import "github.com/mleku/nodl/pkg/codec/kind"

var PrivilegedKinds = &T{[]*kind.T{
	kind.EncryptedDirectMessage,
	kind.GiftWrap,
	kind.GiftWrapWithKind4,
	kind.ApplicationSpecificData,
	kind.Deletion,
}}

func IsPrivileged(k ...*kind.T) (is bool) {
	for i := range PrivilegedKinds.K {
		for j := range k {
			if *(k[j]) == *(PrivilegedKinds.K[i]) {
				return true
			}
		}
	}
	return
}