package util

import "k8s.io/cloud-provider-alibaba-cloud/pkg/model/tag"

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
