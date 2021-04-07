package calculator

import (
	"testing"

	"bigtable-autoscaler.com/m/v2/pkg/pointer"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
)

func TestCalcDesiredNodes(t *testing.T) {
	status := &bigtablev1.BigtableAutoscalerStatus{
		CurrentNodes:          pointer.Int32(1),
		CurrentCPUUtilization: pointer.Int32(30),
	}

	spec := &bigtablev1.BigtableAutoscalerSpec{
		MinNodes:             pointer.Int32(1),
		MaxNodes:             pointer.Int32(10),
		TargetCPUUtilization: pointer.Int32(50),
		MaxScaleDownNodes:    pointer.Int32(5),
	}

	n := CalcDesiredNodes(status, spec)

	t.Errorf("Result: %d", n)
}
