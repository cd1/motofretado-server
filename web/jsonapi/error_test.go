package jsonapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidTypeError_Error(t *testing.T) {
	testCases := []struct {
		name string
		err  InvalidTypeError
	}{
		{
			name: "empty",
		},
		{
			name: "type only",
			err: InvalidTypeError{
				Type: "my-type",
			},
		},
		{
			name: "expected type only",
			err: InvalidTypeError{
				ExpectedType: "my-expected-type",
			},
		},
		{
			name: "complete",
			err: InvalidTypeError{
				Type:         "my-type",
				ExpectedType: "my-expected-type",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			assert.NotEmpty(subT, tc.err.Error())
		})
	}
}

func TestUnsupportedVersionError_Error(t *testing.T) {
	testCases := []struct {
		name string
		err  UnsupportedVersionError
	}{
		{
			name: "empty",
		},
		{
			name: "version only",
			err: UnsupportedVersionError{
				Version: "my-version",
			},
		},
		{
			name: "current version only",
			err: UnsupportedVersionError{
				CurrentVersion: "my-current-version",
			},
		},
		{
			name: "complete",
			err: UnsupportedVersionError{
				Version:        "my-version",
				CurrentVersion: "my-current-version",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(subT *testing.T) {
			assert.NotEmpty(subT, tc.err.Error())
		})
	}
}

func BenchmarkInvalidTypeError_Error(b *testing.B) {
	err := InvalidTypeError{
		Type:         "my-type",
		ExpectedType: "my-expected-type",
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if len(err.Error()) == 0 {
			b.Error(err)
		}
	}
}

func BenchmarkUnsupportedVersionError_Error(b *testing.B) {
	err := UnsupportedVersionError{
		Version:        "my-version",
		CurrentVersion: "my-current-version",
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		if len(err.Error()) == 0 {
			b.Error("empty error message")
		}
	}
}
