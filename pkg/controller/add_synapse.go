package controller

import (
	"github.com/vrutkovs/synapse-operator/pkg/controller/synapse"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, synapse.Add)
}
