package supervisor

import (
	"fmt"
)

func (t *task) setUser() error {
	if t.User != "" {
		return fmt.Errorf("Unsuported opetation for this platform")
	}
	return nil
}
