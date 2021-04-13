package googlecloud

import (
	"context"
	"fmt"

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

func (b *bigtableClientWrapper) Clusters(
	ctx context.Context, instanceID string,
) ([]ClusterInfo, error) {
	clustersInfo, err := b.bigtableClient.Clusters(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to find clusters info for instanceID %s: %w", instanceID, err)
	}

	clustersInfoWrapped := []ClusterInfo{}

	for _, clusterInfo := range clustersInfo {
		clusterInfoWrapped := clusterInfoWrapper{
			clusterInfo: clusterInfo,
		}

		clustersInfoWrapped = append(clustersInfoWrapped, &clusterInfoWrapped)
	}

	return clustersInfoWrapped, nil
}
