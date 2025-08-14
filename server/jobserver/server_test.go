package jobserver

import (
	"log"
	"os"
	"testing"

	"github.com/interline-io/transitland-server/server/jobs/artifactjob"
	"github.com/interline-io/transitland-server/server/testutil"
)

func TestMain(m *testing.M) {
	if a, ok := testutil.CheckTestDB(); !ok {
		log.Print(a)
		return
	}
	os.Exit(m.Run())
}

// TestArtifactJobRegistryInterface ensures the interface is properly defined
func TestArtifactJobRegistryInterface(t *testing.T) {
	// This test ensures the artifact job registry interface is properly imported
	// and can be used in the model configuration
	// The interface is meant to be implemented by deployment-specific registries
	// like the TLV2 River implementation
	_ = artifactjob.JobSubmitter(nil)
}
