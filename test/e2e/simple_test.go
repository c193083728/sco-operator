package e2e

import (
	"testing"

	. "github.com/c193083728/sco-operator/test/support"
)

func TestDesignerDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	_ = test.NewTestNamespace()
}
