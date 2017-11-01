package types

import (
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

// Selector defines SelectorRequirement slice type.
type Selector []SelectorRequirement

// SelectorRequirement is a selector that contains values, a key, and an operator
// that relates the key and values.
type SelectorRequirement struct {
	// The label key that the selector applies to.
	Key string `json:"key"`
	// Represents a key's relationship to a set of values.
	// Valid operators are In, NotIn, Exists, DoesNotExist. Gt, and Lt.
	Operator selection.Operator `json:"operator"`
	// An array of string values. If the operator is In or NotIn,
	// the values array must be non-empty. If the operator is Exists or DoesNotExist,
	// the values array must be empty. If the operator is Gt or Lt, the values
	// array must have a single element.
	Values []string `json:"values,omitempty"`
}

// AsSelector returns selector.
func (n *Selector) AsSelector() labels.Selector {
	requirements := n.AsRequirement()
	return labels.NewSelector().Add(requirements...)
}

// AsRequirement returns requirement.
func (n *Selector) AsRequirement() []labels.Requirement {
	var requirements []labels.Requirement
	for _, sel := range *n {
		requirement, err := labels.NewRequirement(sel.Key, sel.Operator, sel.Values)
		if err != nil {
			log.Infof("selector sel: %v as requirement error: %v", sel, err)
			continue
		}
		requirements = append(requirements, *requirement)
	}
	return requirements
}
