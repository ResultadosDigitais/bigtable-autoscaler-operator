package status

import (
	"context"
	"testing"
	"time"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
	"bigtable-autoscaler.com/m/v2/mocks"
	"bigtable-autoscaler.com/m/v2/pkg/interfaces"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_statusSyncer_Start(t *testing.T) {
	autoscaler := bigtablev1.BigtableAutoscaler{}

	mockStatusWriterWrapper := mocks.WriterWrapper{}
	mockStatusWriterWrapper.On("Update", mock.Anything, &autoscaler).Return(nil)

	cpuUsage := int32(55)
	nodesCount := int32(2)

	mockGoogleCloudClientWrapper := mocks.GoogleCloudClient{}
	mockGoogleCloudClientWrapper.On("GetCurrentCPULoad").Return(cpuUsage, nil)
	mockGoogleCloudClientWrapper.On("GetCurrentNodeCount", "cluster-id").Return(nodesCount, nil)

	type fields struct {
		ctx               context.Context
		statusWriter      interfaces.WriterWrapper
		autoscaler        *bigtablev1.BigtableAutoscaler
		googleCloudClient interfaces.GoogleCloudClient
		clusterID         string
		log               logr.Logger
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "starts the syncer",
			fields: fields{
				ctx:               context.Background(),
				statusWriter:      &mockStatusWriterWrapper,
				autoscaler:        &autoscaler,
				googleCloudClient: &mockGoogleCloudClientWrapper,
				clusterID:         "cluster-id",
				log:               ctrl.Log.WithName("test runtime"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &syncer{
				ctx:               tt.fields.ctx,
				statusWriter:      tt.fields.statusWriter,
				autoscaler:        tt.fields.autoscaler,
				googleCloudClient: tt.fields.googleCloudClient,
				clusterID:         tt.fields.clusterID,
				log:               tt.fields.log,
			}
			s.Start()
		})
	}
	// We need to wait for the go routine
	time.Sleep(1 * time.Millisecond)
	if assert.NotNil(t, autoscaler) {
		assert.Equal(t, int32(55), *autoscaler.Status.CurrentCPUUtilization)
		assert.Equal(t, int32(2), *autoscaler.Status.CurrentNodes)
	}
}
