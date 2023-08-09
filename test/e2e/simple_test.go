package e2e

import (
	"testing"

	. "github.com/sco1237896/sco-operator/test/support"
)

func TestDesignerDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	_ = test.NewTestNamespace()
}
