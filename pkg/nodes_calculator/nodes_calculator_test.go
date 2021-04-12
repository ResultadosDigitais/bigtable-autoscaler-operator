/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nodes_calculator

import (
	"testing"

	"bigtable-autoscaler.com/m/v2/pkg/pointer"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
)

func TestCalcDesiredNodes(t *testing.T) {
	var status *bigtablev1.BigtableAutoscalerStatus
	var spec *bigtablev1.BigtableAutoscalerSpec

	tests := map[string]struct {
		currentNodes int32
		currentCPU   int32
		minNodes     int32
		maxNodes     int32
		targetCPU    int32
		maxScaleDown int32
		expected     int32
	}{
		"scale up":       {currentNodes: 1, currentCPU: 80, minNodes: 1, maxNodes: 10, targetCPU: 50, maxScaleDown: 2, expected: 2},
		"scale up round": {currentNodes: 1, currentCPU: 100, minNodes: 1, maxNodes: 10, targetCPU: 50, maxScaleDown: 2, expected: 2},
		"scale down":     {currentNodes: 10, currentCPU: 5, minNodes: 1, maxNodes: 10, targetCPU: 50, maxScaleDown: 10, expected: 1},
		"max scale down": {currentNodes: 10, currentCPU: 5, minNodes: 1, maxNodes: 10, targetCPU: 50, maxScaleDown: 4, expected: 6},
		"just perfect":   {currentNodes: 5, currentCPU: 50, minNodes: 1, maxNodes: 10, targetCPU: 50, maxScaleDown: 5, expected: 5},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			status = &bigtablev1.BigtableAutoscalerStatus{
				CurrentNodes:          pointer.Int32(test.currentNodes),
				CurrentCPUUtilization: pointer.Int32(test.currentCPU),
			}

			spec = &bigtablev1.BigtableAutoscalerSpec{
				MinNodes:             pointer.Int32(test.minNodes),
				MaxNodes:             pointer.Int32(test.maxNodes),
				TargetCPUUtilization: pointer.Int32(test.targetCPU),
				MaxScaleDownNodes:    pointer.Int32(test.maxScaleDown),
			}

			nodes := CalcDesiredNodes(status, spec)

			if nodes != test.expected {
				t.Errorf("expected: %v, got: %v", test.expected, nodes)
			}
		})
	}
}
