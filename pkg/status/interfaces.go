package status

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Writer interface {
	Update(ctx context.Context, obj runtime.Object, opts ...ctrlclient.UpdateOption) error
}
