package button

import (
	"testing"

	"github.com/tomyan/sumi/runtime/sumitest"
)

func TestButtonSnapshots(t *testing.T) {
	sumitest.AssertSnapshots(t, Scenario())
}
