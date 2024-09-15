/*/
 ** This software is covered by the MIT License.
 ** See: `./LICENSE`.
/*/

package main

import (
	"testing"

	tests "github.com/nacioboi/go_sfda_map/tests"
)

func Test_SFDA(t *testing.T) {
	// Call your benchmark function here for trace
	tests.Test_Consistency(1024) // Example: run your consistency test
}
