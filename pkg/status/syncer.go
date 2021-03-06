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
	"k8s.io/apimachinery/pkg/types"
)

const optimisticLockError = "the object has been modified; please apply your changes to the latest version and try again"
const inexistentResourceError = "invalid object"
const tickTime = 5 * time.Second

type Syncer struct {
	writer  Writer
	running map[types.UID]chan bool
	log     logr.Logger
}

func NewSyncer(writer Writer, log logr.Logger) *Syncer {
	return &Syncer{
		writer:  writer,
		running: make(map[types.UID]chan bool),
		log:     log,
	}
}

func (s *Syncer) Register(
	ctx context.Context,
	autoscaler *bigtablev1.BigtableAutoscaler,
	googleCloudClient googlecloud.GoogleCloudClient,
) {
	if previous_ch, ok := s.running[autoscaler.UID]; ok {
		s.log.Info("Stopping previous routine")
		previous_ch <- true
	}

	ch := make(chan bool, 1)
	s.running[autoscaler.UID] = ch

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		ticker := time.NewTicker(tickTime)
		s.log.Info("Starting new metrics sync routine")

		for {
			select {
			case <-ticker.C:
				currentCpu, err := googleCloudClient.GetCurrentCPULoad()
				if err != nil {
					s.log.Error(err, "failed to get nodes metrics")

					continue
				}
				autoscaler.Status.CurrentCPUUtilization = &currentCpu

				currentNodes, err := googleCloudClient.GetCurrentNodeCount(autoscaler.Spec.BigtableClusterRef.ClusterID)
				if err != nil {
					s.log.Error(err, "failed to get nodes count")

					continue
				}

				autoscaler.Status.CurrentNodes = &currentNodes
				s.log.Info("Metric read", "cpu utilization", currentCpu, "node count", currentNodes, "autoscaler", autoscaler.ObjectMeta.Name)

				if err := s.writer.Update(ctx, autoscaler); err != nil {
					if strings.Contains(err.Error(), inexistentResourceError) {
						s.log.Info("Autoscaler was deleted, stopping syncing.")
						return nil
					}

					if strings.Contains(err.Error(), optimisticLockError) {
						s.log.Error(err, "A minor concurrency error occurred when updating status. We just need to try again.")
						continue
					}

					s.log.Error(err, "failed to update autoscaler status")

					return fmt.Errorf("failed to update autoscaler status: %w", err)
				}

			case <-ch:
				s.log.Info("Interrupted sync from previous version")
				return nil

			}
		}
	})
}
