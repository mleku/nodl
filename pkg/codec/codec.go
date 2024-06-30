// Package codec defines interfaces for use with both pointer and reference
// types that allows proper integration with zero-copy and allocation pools
package codec

type JSON interface {
	// MarshalJSON converts the data of the type into JSON, appending it to the
	// provided slice.
	//
	// The function allows providing a slice to write to so that it can be used
	// without allocating new buffers.
	MarshalJSON(dst B) (b B, err error)
	// UnmarshalJSON decodes a JSON form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	//
	// In order to facilitate the use of reference types (maps and slices) it
	// also must return itself so the updated headers can propagate to the
	// caller.
	UnmarshalJSON(b B) (a any, rem B, err error)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending
	// it to the provided slice.
	//
	// The function allows providing a slice to write to so that it can be used
	// without allocating new buffers.
	MarshalBinary(dst B) (b B)
	// UnmarshalBinary decodes a binary form of a type back into the runtime
	// form, and returns whatever remains after the type has been decoded out.
	//
	// In order to facilitate the use of reference types (maps and slices) it
	// also must return itself so the updated headers can propagate to the
	// caller.
	UnmarshalBinary(b B) (a any, rem B, err error)
}
