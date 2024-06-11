package filters

import "mleku.net/g/nodl/pkg/filter"

type T []filter.T

func (t *T) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

func (t *T) UnmarshalJSON(b []byte) error {
	panic("implement me")
}

func (t *T) MarshalBinary() (data []byte, err error) {
	panic("implement me")
}

func (t *T) UnmarshalBinary(data []byte) error {
	panic("implement me")
}
