package data

import (
	"net/url"
)

type URI struct {
	value  string
	parsed *url.URL
}

func NewURI(value string) (*URI, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return nil, err
	}

	return &URI{
		value:  value,
		parsed: parsed,
	}, nil
}

func NewURIUnsafe(value string) *URI {
	parsed, err := url.Parse(value)
	if err != nil {
		panic(err)
	}
	return &URI{value: value, parsed: parsed}
}

func (u *URI) Value() string {
	return u.value
}

func (u *URI) String() string {
	return u.value
}

func (u *URI) Scheme() string {
	return u.parsed.Scheme
}

func (u *URI) Host() string {
	return u.parsed.Host
}

func (u *URI) Path() string {
	return u.parsed.Path
}

func (u *URI) Equal(other *URI) bool {
	return u.Value() == other.Value()
}
