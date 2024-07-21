// Package codec defines interfaces for use with both pointer and reference
// types that allows proper integration with zero-copy and allocation pools
package codec

type JSON interface {
	// MarshalJSON converts the data of the type into JSON, appending it to the
	// provided slice and returning the extended slice.
	MarshalJSON(dst B) (b B, err error)
	// UnmarshalJSON decodes a JSON form of a type back into the runtime form,
	// and returns whatever remains after the type has been decoded out.
	UnmarshalJSON(b B) (r B, err error)
}

type Binary interface {
	// MarshalBinary converts the data of the type into binary form, appending
	// it to the provided slice.
	MarshalBinary(dst B) (b B)
	// UnmarshalBinary decodes a binary form of a type back into the runtime
	// form, and returns whatever remains after the type has been decoded out.
	UnmarshalBinary(b B) (r B, err error)
}
