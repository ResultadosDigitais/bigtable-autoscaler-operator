package status

import (
	"context"
	"fmt"
	"strings"
	"time"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
	"bigtable-autoscaler.com/m/v2/pkg/googlecloud"
	"github.com/go-logr/logr"
	"golang.org/x/sync/errgroup"
)

const optimisticLockError = "the object has been modified; please apply your changes to the latest version and try again"
const inexistentResourceError = "invalid object"
const tickTime = 3 * time.Second

type Syncer struct {
	writer Writer
	log    logr.Logger
}

type syncerInstance struct {
}

// syncers: make(map[types.NamespacedName]*status.Syncer),

// TODO: register new specs to sync
// TODO: remove specs from sync list

func NewSyncer(writer Writer, log logr.Logger) *Syncer {
	return &Syncer{
		writer: writer,
		log:    log,
	}
}

func (s *Syncer) Start(
	ctx context.Context,
	autoscaler *bigtablev1.BigtableAutoscaler,
	googleCloudClient googlecloud.GoogleCloudClient,
) {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		ticker := time.NewTicker(tickTime)
		for ; true; <-ticker.C {
			currentCpu, err := googleCloudClient.GetCurrentCPULoad()
			if err != nil {
				return fmt.Errorf("failed to get metrics: %w", err)
			}
			autoscaler.Status.CurrentCPUUtilization = &currentCpu

			currentNodes, err := googleCloudClient.GetCurrentNodeCount(autoscaler.Spec.BigtableClusterRef.ClusterID)
			if err != nil {
				s.log.Error(err, "failed to get nodes count")

				return fmt.Errorf("failed to get nodes count: %w", err)
			}

			autoscaler.Status.CurrentNodes = &currentNodes
			s.log.Info("Metric read", "cpu utilization", currentCpu, "node count", currentNodes, "autoscaler", autoscaler.ObjectMeta.Name)

			if err := s.writer.Update(ctx, autoscaler); err != nil {
				if strings.Contains(err.Error(), inexistentResourceError) {
					s.log.Info("Resource not found")
					break
				}

				if strings.Contains(err.Error(), optimisticLockError) {
					s.log.Error(err, "A minor concurrency error occurred when updating status. We just need to try again.")
					continue
				}

				s.log.Error(err, "failed to update autoscaler status")

				return fmt.Errorf("failed to update autoscaler status: %w", err)
			}
		}

		return nil
	})
}
