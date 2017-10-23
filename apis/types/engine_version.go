// Code generated by go-swagger; DO NOT EDIT.

package types

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// EngineVersion engine version
// swagger:model EngineVersion

type EngineVersion struct {

	// Api Version held by Pouchd
	APIVersion string `json:"ApiVersion,omitempty"`

	// arch
	Arch string `json:"Arch,omitempty"`

	// build time
	BuildTime string `json:"BuildTime,omitempty"`

	// experimental
	Experimental bool `json:"Experimental,omitempty"`

	// Commit ID held by the latest commit operation
	GitCommit string `json:"GitCommit,omitempty"`

	// go version
	GoVersion string `json:"GoVersion,omitempty"`

	// kernel version
	KernelVersion string `json:"KernelVersion,omitempty"`

	// os
	Os string `json:"Os,omitempty"`

	// version
	Version string `json:"Version,omitempty"`
}

/* polymorph EngineVersion ApiVersion false */

/* polymorph EngineVersion Arch false */

/* polymorph EngineVersion BuildTime false */

/* polymorph EngineVersion Experimental false */

/* polymorph EngineVersion GitCommit false */

/* polymorph EngineVersion GoVersion false */

/* polymorph EngineVersion KernelVersion false */

/* polymorph EngineVersion Os false */

/* polymorph EngineVersion Version false */

// Validate validates this engine version
func (m *EngineVersion) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *EngineVersion) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *EngineVersion) UnmarshalBinary(b []byte) error {
	var res EngineVersion
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
