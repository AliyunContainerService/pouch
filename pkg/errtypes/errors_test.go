package errtypes

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCheckError(t *testing.T) {
	if !checkError(errors.Wrap(ErrNotfound, "test"), codeNotfound) {
		t.Error("check Wrap error")
	}

	if !checkError(errors.Wrapf(ErrNotfound, "test"), codeNotfound) {
		t.Error("check Wrapf error")
	}

	if !checkError(errors.WithMessage(ErrNotfound, "test"), codeNotfound) {
		t.Error("check WithMessage error")
	}

	if !checkError(errors.Wrap(errors.Wrap(ErrNotfound, "test"), "test2"), codeNotfound) {
		t.Error("check Wrap error")
	}
}
