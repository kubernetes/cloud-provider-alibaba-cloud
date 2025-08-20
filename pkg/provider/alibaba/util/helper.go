package util

import (
	"fmt"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"strings"
)

// A PaginationResponse represents a response with pagination information
type PaginationResult struct {
	TotalCount int
	PageNumber int
	PageSize   int
}

type Pagination struct {
	PageNumber int
	PageSize   int
}

// NextPage gets the next page of the result set
func (r *PaginationResult) NextPage() *Pagination {
	if r.PageNumber*r.PageSize >= r.TotalCount {
		return nil
	}
	return &Pagination{PageNumber: r.PageNumber + 1, PageSize: r.PageSize}
}

// providerID
// 1) the id of the instance in the alicloud API. Use '.' to separate providerID which looks like 'cn-hangzhou.i-v98dklsmnxkkgiiil7'. The format of "REGION.NODEID"
// 2) the id for an instance in the kubernetes API, which has 'alicloud://' prefix. e.g. alicloud://cn-hangzhou.i-v98dklsmnxkkgiiil7
func NodeFromProviderID(providerID string) (string, string, error) {
	if strings.HasPrefix(providerID, "alicloud://") {
		k8sName := strings.Split(providerID, "://")
		if len(k8sName) < 2 {
			return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
		} else {
			providerID = k8sName[1]
		}
	}

	name := strings.Split(providerID, ".")
	if len(name) < 2 {
		return "", "", fmt.Errorf("alicloud: unable to split instanceid and region from providerID, error unexpected providerID=%s", providerID)
	}
	return name[0], name[1], nil
}

func ProviderIDFromInstance(region, instance string) string {
	return fmt.Sprintf("%s.%s", region, instance)
}

func SDKError(api string, err error) error {
	if err == nil {
		return err
	}
	switch err := err.(type) {
	case *tea.SDKError:
		if err == nil || err.Message == nil {
			return err
		}
		attr := strings.Split(tea.StringValue(err.Message), "request id:")
		if len(attr) < 2 {
			return err
		}
		err.SetErrMsg(fmt.Sprintf("[SDKError] API: %s,StatusCode: %d, ErrorCode: %s, RequestId: %s, Message: %s",
			api, tea.IntValue(err.StatusCode), tea.StringValue(err.Code), attr[1], attr[0]))
		return err
	case *errors.ServerError:
		return fmt.Errorf("[SDKError] API: %s, ErrorCode: %s, RequestId: %s, Message: %s",
			api, err.ErrorCode(), err.RequestId(), err.Message())
	case *errors.ClientError:
		return err.OriginError()
	default:
		return err
	}
}
