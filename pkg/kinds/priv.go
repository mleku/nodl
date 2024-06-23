package kinds

import "github.com/mleku/nodl/pkg/kind"

var PrivilegedKinds = T{
	kind.EncryptedDirectMessage,
	kind.GiftWrap,
	kind.GiftWrapWithKind4,
	kind.ApplicationSpecificData,
	kind.Deletion,
}

func IsPrivileged(k ...kind.T) (is bool) {
	for i := range PrivilegedKinds {
		for j := range k {
			if k[j] == PrivilegedKinds[i] {
				return true
			}
		}
	}
	return
}