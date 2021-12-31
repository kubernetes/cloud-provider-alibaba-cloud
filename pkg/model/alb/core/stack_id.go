package core

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

type StackID types.NamespacedName

func (stackID StackID) String() string {
	if stackID.Namespace == "" {
		return stackID.Name
	}
	return fmt.Sprintf("%s/%s", stackID.Namespace, stackID.Name)
}
