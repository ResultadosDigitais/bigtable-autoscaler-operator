package status

import (
	"context"

	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type writerWrapper struct {
	statusWriter ctrlclient.StatusWriter
}

func (s *writerWrapper) Update(ctx context.Context, obj runtime.Object, opts ...ctrlclient.UpdateOption) error {
	return s.Update(ctx, obj, opts...)
}

// Make sure the wrapper complies with its interface.
var _ (interfaces.WriterWrapper) = (*writerWrapper)(nil)
