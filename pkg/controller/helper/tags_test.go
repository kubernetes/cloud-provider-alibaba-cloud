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

func TestDiffLoadBalancerTags(t *testing.T) {
	tests := []struct {
		name       string
		local      []tag.Tag
		remote     []tag.Tag
		wantTag    []tag.Tag
		wantUntag  []tag.Tag
	}{
		{
			name: "add new tags",
			local: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			remote: []tag.Tag{},
			wantTag: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			wantUntag: []tag.Tag{},
		},
		{
			name:  "remove tags",
			local: []tag.Tag{},
			remote: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			wantTag: []tag.Tag{},
			wantUntag: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
		},
		{
			name: "update existing tags",
			local: []tag.Tag{
				{Key: "key1", Value: "value1-new"},
				{Key: "key2", Value: "value2"},
			},
			remote: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
				{Key: "key3", Value: "value3"},
			},
			wantTag: []tag.Tag{
				{Key: "key1", Value: "value1-new"},
			},
			wantUntag: []tag.Tag{
				{Key: "key3", Value: "value3"},
			},
		},
		{
			name: "no changes",
			local: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			remote: []tag.Tag{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
			},
			wantTag:   []tag.Tag{},
			wantUntag: []tag.Tag{},
		},
		{
			name: "mixed operations",
			local: []tag.Tag{
				{Key: "key1", Value: "value1"}, // unchanged
				{Key: "key2", Value: "value2-new"}, // updated
				{Key: "key4", Value: "value4"}, // added
			},
			remote: []tag.Tag{
				{Key: "key1", Value: "value1"}, // unchanged
				{Key: "key2", Value: "value2"}, // updated
				{Key: "key3", Value: "value3"}, // removed
			},
			wantTag: []tag.Tag{
				{Key: "key2", Value: "value2-new"},
				{Key: "key4", Value: "value4"},
			},
			wantUntag: []tag.Tag{
				{Key: "key3", Value: "value3"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTag, gotUntag := DiffLoadBalancerTags(tt.local, tt.remote)
			assert.ElementsMatch(t, tt.wantTag, gotTag, "tag operations mismatch")
			assert.ElementsMatch(t, tt.wantUntag, gotUntag, "untag operations mismatch")
		})
	}
}
