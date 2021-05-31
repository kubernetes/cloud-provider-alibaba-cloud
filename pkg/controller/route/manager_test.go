package route

import (
	"context"
	"github.com/stretchr/testify/assert"
	globalCtx "k8s.io/cloud-provider-alibaba-cloud/pkg/context"
	"k8s.io/cloud-provider-alibaba-cloud/pkg/provider/vmock"
	"testing"
)

func TestGetRouteTables(t *testing.T) {
	meta := vmock.NewMockMetaData("vpc-xxx")
	noRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil, nil), IMetaData: meta}
	singleRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil, []string{"table-bbb"}), IMetaData: meta}
	multiRouteTableVPC := vmock.MockCloud{MockVPC: vmock.NewMockVPC(nil, []string{"table-ccc", "table-ddd"}), IMetaData: meta}

	globalCtx.CFG.Global.RouteTableIDS = "table-xxx,table-yyy"
	tables, err := getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 2, len(tables), "assert route tables from cloud config")
	assert.Equal(t, nil, err, "assert route tables from cloud config")

	globalCtx.CFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), noRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from no table vpc")
	assert.Equal(t, false, err == nil, "assert route tables from no table vpc")

	globalCtx.CFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), singleRouteTableVPC)
	assert.Equal(t, 1, len(tables), "assert route tables from no single vpc")
	assert.Equal(t, nil, err, "assert route tables from single table vpc")

	globalCtx.CFG.Global.RouteTableIDS = ""
	tables, err = getRouteTables(context.Background(), multiRouteTableVPC)
	assert.Equal(t, 0, len(tables), "assert route tables from multi table vpc")
	assert.Equal(t, false, err == nil, "assert route tables from multi table vpc")
}
