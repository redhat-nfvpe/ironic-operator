package controller

import (
	"github.com/redhat-nfvpe/ironic-operator/pkg/controller/ironicconductor"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, ironicconductor.Add)
}
