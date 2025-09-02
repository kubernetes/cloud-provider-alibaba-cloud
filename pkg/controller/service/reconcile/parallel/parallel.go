package parallel

import (
	"context"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/util/workqueue"
	"sync"
)

const (
	DefaultWorkerCount = 10
)

func DoPiece(ctx context.Context, workers, taskLens int, fn func(i int)) {
	if workers == 0 {
		workers = DefaultWorkerCount
	}
	workqueue.ParallelizeUntil(ctx, workers, taskLens, fn)
}

func Do(fns ...func() error) error {
	count := len(fns)
	errs := make([]error, count)
	wg := sync.WaitGroup{}
	for i, fn := range fns {
		wg.Add(1)
		go func(i int, fn func() error) {
			defer wg.Done()
			errs[i] = fn()
		}(i, fn)
	}
	wg.Wait()
	return utilerrors.NewAggregate(errs)
}
