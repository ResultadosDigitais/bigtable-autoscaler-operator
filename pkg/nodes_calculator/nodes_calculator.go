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
	"math"

	bigtablev1 "bigtable-autoscaler.com/m/v2/api/v1"
)

func CalcDesiredNodes(status *bigtablev1.BigtableAutoscalerStatus, spec *bigtablev1.BigtableAutoscalerSpec) int32 {
	currentNodes := *status.CurrentNodes
	totalCPU := *status.CurrentCPUUtilization * currentNodes
	desiredNodes := totalCPU / *spec.TargetCPUUtilization

	if (currentNodes - desiredNodes) > *spec.MaxScaleDownNodes {
		desiredNodes = currentNodes - *spec.MaxScaleDownNodes
	}

	desiredNodes = int32(math.Max(float64(desiredNodes), float64(*spec.MinNodes)))
	desiredNodes = int32(math.Min(float64(desiredNodes), float64(*spec.MaxNodes)))

	return desiredNodes
}
