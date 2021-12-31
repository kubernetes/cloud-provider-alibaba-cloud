package core

// Resource represents a deployment unit.
type Resource interface {
	// resource's stack.
	Stack() Manager

	// resource's Type.
	Type() string

	// resource's ID within stack.
	ID() string
}

// NewResourceMeta constructs new resource metadata.
func NewResourceMeta(stack Manager, resType string, id string) ResourceMeta {
	return ResourceMeta{
		stack:   stack,
		resType: resType,
		id:      id,
	}
}

// Metadata for all resources.
type ResourceMeta struct {
	stack   Manager
	resType string
	id      string
}

func (m *ResourceMeta) Stack() Manager {
	return m.stack
}

func (m *ResourceMeta) Type() string {
	return m.resType
}

func (m *ResourceMeta) ID() string {
	return m.id
}

// ResourceVisitor represents a functor that can operate on a resource.
type ResourceVisitor interface {
	Visit(res Resource) error
}
