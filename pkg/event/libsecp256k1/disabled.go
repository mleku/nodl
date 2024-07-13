//go:build !libsecp256k1

package libsecp256k1

func Sign(msg [32]byte, sk [32]byte) ([64]byte, error) {
	panic("libsecp256k1 not enabled in build")
}
func Verify(msg [32]byte, sig [64]byte, pk [32]byte) bool {
	panic("libsecp256k1 not enabled in build")
}
