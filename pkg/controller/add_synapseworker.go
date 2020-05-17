package controller

import (
	"github.com/vrutkovs/synapse-operator/pkg/controller/synapseworker"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, synapseworker.Add)
}
