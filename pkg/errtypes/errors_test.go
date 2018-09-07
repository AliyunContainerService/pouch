package errtypes

import (
	"testing"

	"github.com/pkg/errors"
)

func TestCheckError(t *testing.T) {
	if !checkError(errors.Wrap(ErrNotfound, "test"), codeNotFound) {
		t.Error("check Wrap error")
	}

	if !checkError(errors.Wrapf(ErrNotfound, "test"), codeNotFound) {
		t.Error("check Wrapf error")
	}

	if !checkError(errors.WithMessage(ErrNotfound, "test"), codeNotFound) {
		t.Error("check WithMessage error")
	}

	if !checkError(errors.Wrap(errors.Wrap(ErrNotfound, "test"), "test2"), codeNotFound) {
		t.Error("check Wrap error")
	}
}
