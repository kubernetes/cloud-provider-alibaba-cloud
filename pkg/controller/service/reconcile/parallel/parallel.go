package parallel

import (
	"context"
	"k8s.io/client-go/util/workqueue"
)

const (
	DefaultWorkerCount = 10
)

func Parallelize(ctx context.Context, workers, taskLens int, fn func(i int)) {
	if workers == 0 {
		workers = DefaultWorkerCount
	}
	workqueue.ParallelizeUntil(ctx, workers, taskLens, fn)
}
