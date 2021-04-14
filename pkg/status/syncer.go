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
	ctx               context.Context
	writer            Writer
	autoscaler        *bigtablev1.BigtableAutoscaler
	googleCloudClient googlecloud.GoogleCloudClient
	clusterID         string
	log               logr.Logger
}

// TODO: register new specs to sync
// TODO: remove specs from sync list

func NewSyncer(
	ctx context.Context,
	writer Writer,
	autoscaler *bigtablev1.BigtableAutoscaler,
	googleCloudClient googlecloud.GoogleCloudClient,
	clusterID string,
	log logr.Logger,
) *Syncer {
	if autoscaler.Status.CurrentCPUUtilization == nil {
		var cpuUsage int32
		autoscaler.Status.CurrentCPUUtilization = &cpuUsage
	}

	return &Syncer{
		ctx:               ctx,
		writer:            writer,
		autoscaler:        autoscaler,
		googleCloudClient: googleCloudClient,
		clusterID:         clusterID,
		log:               log,
	}
}

func (s *Syncer) Start() {
	eg, ctx := errgroup.WithContext(s.ctx)

	eg.Go(func() error {
		ticker := time.NewTicker(tickTime)
		for ; true; <-ticker.C {
			currentCpu, err := s.googleCloudClient.GetCurrentCPULoad()
			if err != nil {
				return fmt.Errorf("failed to get metrics: %w", err)
			}
			s.autoscaler.Status.CurrentCPUUtilization = &currentCpu

			currentNodes, err := s.googleCloudClient.GetCurrentNodeCount(s.clusterID)
			if err != nil {
				s.log.Error(err, "failed to get nodes count")

				return fmt.Errorf("failed to get nodes count: %w", err)
			}

			s.autoscaler.Status.CurrentNodes = &currentNodes
			s.log.Info("Metric read", "cpu utilization", currentCpu, "node count", currentNodes, "autoscaler", s.autoscaler.ObjectMeta.Name)

			if err := s.writer.Update(ctx, s.autoscaler); err != nil {
				if strings.Contains(err.Error(), inexistentResourceError) {
					s.log.Info("Resource not found")
					break
				}

				if strings.Contains(err.Error(), optimisticLockError) {
					s.log.Info("A minor concurrency error occurred when updating status. We just need to try again.")
					continue
				}

				s.log.Error(err, "failed to update autoscaler status")

				return fmt.Errorf("failed to update autoscaler status: %w", err)
			}
		}

		return nil
	})
}
