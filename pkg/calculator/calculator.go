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

package calculator

func CalcDesiredNodes(currentCPU, currentNodes, targetCPU, minNodes, maxNodes, maxScaleDownNodes int32) int32 {
	totalCPU := currentCPU * currentNodes
	desiredNodes := totalCPU/targetCPU + 1 // roundup

	if (currentNodes - desiredNodes) > maxScaleDownNodes {
		desiredNodes = currentNodes - maxScaleDownNodes
	}

	switch {
	case desiredNodes < minNodes:
		return minNodes

	case desiredNodes > maxNodes:
		return maxNodes

	default:
		return desiredNodes
	}
}
