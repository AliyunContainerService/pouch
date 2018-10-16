package ctrd

import (
	"fmt"
	"testing"

	"github.com/alibaba/pouch/pkg/errtypes"

	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
)

func Test_convertCtrdErr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		returnedErr error
	}{
		{
			name: "nil",
			args: args{
				err: nil,
			},
			wantErr:     false,
			returnedErr: nil,
		},
		{
			name: "not found",
			args: args{
				err: errors.Wrap(errdefs.ErrNotFound, "container asdfghjk"),
			},
			wantErr:     true,
			returnedErr: errors.Wrap(errtypes.ErrNotfound, errors.Wrap(errdefs.ErrNotFound, "container asdfghjk").Error()),
		},
		{
			name: "invalid params",
			args: args{
				err: errors.Wrap(errdefs.ErrInvalidArgument, "container asdfghjk"),
			},
			wantErr:     true,
			returnedErr: errors.Wrap(errtypes.ErrInvalidParam, errors.Wrap(errdefs.ErrInvalidArgument, "container asdfghjk").Error()),
		},
		{
			name: "already exists",
			args: args{
				err: errors.Wrap(errdefs.ErrAlreadyExists, "container asdfghjk"),
			},
			wantErr:     true,
			returnedErr: errors.Wrap(errtypes.ErrAlreadyExisted, errors.Wrap(errdefs.ErrAlreadyExists, "container asdfghjk").Error()),
		},
		{
			name: "not implemented",
			args: args{
				err: errors.Wrap(errdefs.ErrNotImplemented, "container asdfghjk"),
			},
			wantErr:     true,
			returnedErr: errors.Wrap(errtypes.ErrNotImplemented, errors.Wrap(errdefs.ErrNotImplemented, "container asdfghjk").Error()),
		},
		{
			name: "normal error",
			args: args{
				err: fmt.Errorf("this is a normal error"),
			},
			wantErr:     true,
			returnedErr: fmt.Errorf("this is a normal error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := convertCtrdErr(tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertCtrdErr() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && (err.Error() != tt.returnedErr.Error()) {
				t.Errorf("convertCtrdErr() error = %v, wantErr %v, returnedErr: %v", err, tt.wantErr, tt.returnedErr)
			}
		})
	}
}

func TestLockFailedError(t *testing.T) {
	type args struct {
		containerID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal test case",
			args: args{
				containerID: "asdfghjkl",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LockFailedError(tt.args.containerID); (err != nil) != tt.wantErr {
				t.Errorf("LockFailedError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
