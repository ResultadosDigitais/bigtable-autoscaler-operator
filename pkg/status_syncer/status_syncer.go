package status_syncer

import (
	"context"
	"fmt"
	"strings"
	"time"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const optimisticLockErrorMsg = "the object has been modified; please apply your changes to the latest version and try again"

type statusSyncer struct {
	ctrlClient        ctrlclient.Client
	autoscaler bigtablev1.BigtableAutoscaler
	googleCloudClient interfaces.GoogleCloudClient
	clusterID         string
	ctx               context.Context
	log               logr.Logger
}

func NewStatusSyncer(
	ctrlClient ctrlclient.Client,
	autoscaler bigtablev1.BigtableAutoscaler,
	googleCloundClient interfaces.GoogleCloudClient, clusterID string, ctx context.Context, log logr.Logger,
) (*statusSyncer, error) {
	return &statusSyncer{
		ctrlClient:        ctrlClient,
		autoscaler:        autoscaler,
		googleCloudClient: googleCloundClient,
		clusterID:         clusterID,
		ctx:               ctx,
		log:               log,
	}, nil
}

func (s *statusSyncer) SyncStatus() {
	eg, ctx := errgroup.WithContext(s.ctx)

	eg.Go(func() error {
		ticker := time.NewTicker(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				var err error
				metric, err := s.googleCloudClient.GetCurrentCPULoad()
				if err != nil {
					return fmt.Errorf("failed to get metrics: %w", err)
				}

				s.log.V(1).Info("Metric read", "cpu utilization", metric)
				s.autoscaler.Status.CurrentCPUUtilization = &metric

				if err = s.ctrlClient.Status().Update(ctx, &s.autoscaler); err != nil {
					if strings.Contains(err.Error(), optimisticLockErrorMsg) {
						s.log.Info("A minor concurrency error occurred when updating status. We just need to try again.")
						continue
					}
					s.log.Error(err, "failed to update autoscaler status")
					return fmt.Errorf("failed to update autoscaler status: %w", err)
				}
			}
		}
	})
}
