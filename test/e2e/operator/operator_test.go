package operator

import (
	"testing"

	. "github.com/sco1237896/sco-operator/test/support"
)

func TestOperatorDeploy(t *testing.T) {
	test := With(t)
	test.T().Parallel()

	_ = test.NewTestNamespace()
}
