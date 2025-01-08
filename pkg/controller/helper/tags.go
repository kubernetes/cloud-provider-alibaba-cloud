package helper

import (
	"k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"
	"strings"
)

func DiffLoadBalancerTags(local, remote []tag.Tag) ([]tag.Tag, []tag.Tag) {
	var needTag, needUntag []tag.Tag

	remoteMap := map[string]string{}
	localMap := map[string]string{}
	for _, r := range remote {
		remoteMap[r.Key] = r.Value
	}

	for _, l := range local {
		localMap[l.Key] = l.Value
		if k, ok := remoteMap[l.Key]; ok && k == l.Value {
			continue
		}
		needTag = append(needTag, l)
	}

	for _, r := range remote {
		if _, ok := localMap[r.Key]; !ok {
			needUntag = append(needUntag, r)
		}
	}

	return needTag, needUntag
}

func FilterTags(tags, defaultTags []tag.Tag, isUserManaged bool) []tag.Tag {
	var filteredTags []tag.Tag

	if isUserManaged {
		defaultTags = append(defaultTags, tag.Tag{Key: REUSEKEY, Value: "true"})
	}

	for _, r := range tags {
		// ignore system tags
		if strings.HasPrefix(r.Key, "acs:") {
			continue
		}

		// ignore default tags
		found := false
		for _, d := range defaultTags {
			if r.Key == d.Key {
				found = true
				break
			}
		}

		if !found {
			filteredTags = append(filteredTags, r)
		}
	}

	return filteredTags
}
