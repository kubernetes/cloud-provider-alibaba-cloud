package albconfigmanager

import "k8s.io/cloud-provider-alibaba-cloud/pkg/model/alb/core"

type StackSchema struct {
	ID string `json:"id"`

	Resources map[string]map[string]interface{} `json:"resources"`
}

func NewStackSchemaBuilder(stackID core.StackID) *stackSchemaBuilder {
	return &stackSchemaBuilder{
		stackID:   stackID,
		resources: make(map[string]map[string]interface{}),
	}
}

var _ core.ResourceVisitor = &stackSchemaBuilder{}

type stackSchemaBuilder struct {
	stackID   core.StackID
	resources map[string]map[string]interface{}
}

func (b *stackSchemaBuilder) Visit(res core.Resource) error {
	if _, ok := b.resources[res.Type()]; !ok {
		b.resources[res.Type()] = make(map[string]interface{})
	}
	b.resources[res.Type()][res.ID()] = res
	return nil
}

func (b *stackSchemaBuilder) Build() StackSchema {
	return StackSchema{
		ID:        b.stackID.String(),
		Resources: b.resources,
	}
}
