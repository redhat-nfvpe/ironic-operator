package controller

import (
	"github.com/metalkube/ironic-operator/pkg/controller/ironic"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, ironic.Add)
}
