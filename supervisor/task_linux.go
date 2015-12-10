package supervisor

import (
	"fmt"
	"os/user"
	"strconv"
	"syscall"
)

func (t *task) setUser() error {
	if t.User == "" {
		return nil
	}
	if u, err := user.Lookup(t.User); err == nil {
		if t.cmd.SysProcAttr == nil {
			t.cmd.SysProcAttr = &syscall.SysProcAttr{}
			t.cmd.SysProcAttr.Credential = &syscall.Credential{}
		}
		if uid, err := strconv.ParseUint(u.Uid, 10, 32); err == nil {
			t.cmd.SysProcAttr.Credential.Uid = uint32(uid)
		}
		if gid, err := strconv.ParseUint(u.Gid, 10, 32); err == nil {
			t.cmd.SysProcAttr.Credential.Gid = uint32(gid)
		}
	} else {
		return fmt.Errorf("While lookup user %s for task %s: %s\n", t.User, t.name, err.Error())
	}
	return nil
}
