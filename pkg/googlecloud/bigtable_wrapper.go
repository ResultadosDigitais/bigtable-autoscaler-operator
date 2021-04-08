package googlecloud

import (
	"context"
	"errors"
	"fmt"

	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"cloud.google.com/go/bigtable"
)

type bigtableClientWrapper struct {
	bigtableClient *bigtable.InstanceAdminClient
}

type clustersInfoWrapper struct {
	clustersInfo []*bigtable.ClusterInfo
}

func (c *clustersInfoWrapper) NodesOfInstance(clusterID string) (int32, error) {
	for _, clusterInfo := range c.clustersInfo {
		if clusterInfo.Name == clusterID {
			return int32(clusterInfo.ServeNodes), nil
		}
	}
	message := fmt.Sprintf("Cluster of id %s not found", clusterID)
	return -1, errors.New(message)
}

// Make sure the wrapper complies with its interface
var _ interfaces.ClustersInfoWrapper = (*clustersInfoWrapper)(nil)

func (b *bigtableClientWrapper) Clusters(ctx context.Context, instanceID string) (interfaces.ClustersInfoWrapper, error) {
	clustersInfo, err := b.bigtableClient.Clusters(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	clustersInfoWrapped := clustersInfoWrapper{
		clustersInfo: clustersInfo,
	}

	return &clustersInfoWrapped, nil
}

// Make sure the wrapper complies with its interface
var _ interfaces.BigtableClientWrapper = (*bigtableClientWrapper)(nil)
