package googlecloud

import (
	"context"
	"fmt"

	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"cloud.google.com/go/bigtable"
)

type bigtableClientWrapper struct {
	bigtableClient *bigtable.InstanceAdminClient
}

type clusterInfoWrapper struct {
	clusterInfo *bigtable.ClusterInfo
}

func (c *clusterInfoWrapper) Name() string {
	return c.clusterInfo.Name
}

func (c *clusterInfoWrapper) ServerNodes() int32 {
	return int32(c.clusterInfo.ServeNodes)
}

// Make sure the wrapper complies with its interface.
var _ interfaces.ClusterInfoWrapper = (*clusterInfoWrapper)(nil)

func (b *bigtableClientWrapper) Clusters(ctx context.Context, instanceID string) ([]interfaces.ClusterInfoWrapper, error) {
	clustersInfo, err := b.bigtableClient.Clusters(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("Failed to find clusters info for instanceID %s: %v", instanceID, err)
	}

	clustersInfoWrapped := []interfaces.ClusterInfoWrapper{}

	for _, clusterInfo := range clustersInfo {
		clusterInfoWrapped := clusterInfoWrapper{
			clusterInfo: clusterInfo,
		}

		clustersInfoWrapped = append(clustersInfoWrapped, &clusterInfoWrapped)
	}

	return clustersInfoWrapped, nil
}

// Make sure the wrapper complies with its interface.
var _ interfaces.BigtableClientWrapper = (*bigtableClientWrapper)(nil)
