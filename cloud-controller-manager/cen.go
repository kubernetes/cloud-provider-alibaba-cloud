package alicloud

import (
	"context"
	"github.com/denverdino/aliyungo/cen"
)

type CEN interface {
	PublishRouteEntries(ctx context.Context, args *cen.PublishRouteEntriesArgs) error
	DescribePublishedRouteEntries(
		ctx context.Context,
		args *cen.DescribePublishedRouteEntriesArgs,
	) (response *cen.DescribePublishedRouteEntriesResponse, err error)
}
