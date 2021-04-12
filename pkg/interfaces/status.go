package interfaces

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)


type WriterWrapper interface {
	Update(ctx context.Context, obj runtime.Object, opts ...ctrlclient.UpdateOption) error
}
