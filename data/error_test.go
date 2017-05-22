package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDuplicateError_Error(t *testing.T) {
	testCases := []struct {
		name string
		err  DuplicateError
	}{
		{
			name: "empty",
		},
		{
			name: "complete",
			err:  DuplicateError{"foo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			assert.NotEmpty(subT, tc.err.Error())
		})
	}
}

func TestInvalidParameterError_Error(t *testing.T) {
	testCases := []struct {
		name string
		err  InvalidParameterError
	}{
		{
			name: "empty",
		},
		{
			name: "name-only",
			err:  InvalidParameterError{Name: "foo"},
		},
		{
			name: "value-only",
			err:  InvalidParameterError{Name: "foo"},
		},
		{
			name: "complete",
			err: InvalidParameterError{
				Name:  "foo",
				Value: "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			assert.NotEmpty(subT, tc.err.Error())
		})
	}
}

func TestMissingParameterError_Error(t *testing.T) {
	testCases := []struct {
		name string
		err  MissingParameterError
	}{
		{
			name: "empty",
		},
		{
			name: "complete",
			err:  MissingParameterError{Name: "foo"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			assert.NotEmpty(subT, tc.err.Error())
		})
	}
}

func BenchmarkDuplicateError_Error(b *testing.B) {
	err := DuplicateError{"foo"}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = err.Error()
	}
}

func BenchmarkInvalidParameterError_Error(b *testing.B) {
	err := InvalidParameterError{
		Name:  "foo",
		Value: "bar",
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = err.Error()
	}
}

func BenchmarkMissingParameterError_Error(b *testing.B) {
	err := MissingParameterError{Name: "foo"}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		_ = err.Error()
	}
}
