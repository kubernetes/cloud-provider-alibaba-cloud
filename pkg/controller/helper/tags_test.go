package helper

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/alibaba/base"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/util"
	"testing"
)

func TestFilterRemoteTags(t *testing.T) {
	tests := []struct {
		name           string
		tags           []tag.Tag
		isUserManaged  bool
		expectedResult []tag.Tag
	}{
		{
			name: "ignore system tags",
			tags: []tag.Tag{
				{Key: "acs:tag:createdby", Value: "value"},
				{Key: "acs:systemTag", Value: "value"},
				{Key: "userTag", Value: "value"},
			},
			isUserManaged: false,
			expectedResult: []tag.Tag{
				{Key: "userTag", Value: "value"},
			},
		},
		{
			name: "ignore default tags without reused tag",
			tags: []tag.Tag{
				{Key: TAGKEY, Value: "a123456"},
				{Key: util.ClusterTagKey, Value: base.CLUSTER_ID},
				{Key: REUSEKEY, Value: "true"},
				{Key: "userTag", Value: "value"},
			},
			isUserManaged: false,
			expectedResult: []tag.Tag{
				{Key: REUSEKEY, Value: "true"},
				{Key: "userTag", Value: "value"},
			},
		},
		{
			name: "ignore default tags with reused tag",
			tags: []tag.Tag{
				{Key: TAGKEY, Value: "a123456"},
				{Key: util.ClusterTagKey, Value: base.CLUSTER_ID},
				{Key: REUSEKEY, Value: "true"},
				{Key: "userTag", Value: "value"},
			},
			isUserManaged: true,
			expectedResult: []tag.Tag{
				{Key: "userTag", Value: "value"},
			},
		},
	}

	defaultTags := []tag.Tag{
		{Key: TAGKEY, Value: "a123456"},
		{Key: util.ClusterTagKey, Value: base.CLUSTER_ID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := FilterTags(test.tags, defaultTags, test.isUserManaged)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}
