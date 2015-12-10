package supervisor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type task struct {
	name string
	cmd  *exec.Cmd

	log     Log
	Command string `json:"command"`
	User    string `json:"user"`
	WorkDir string `json:"work_dir"`
	Env     string `json:"env"`
	Args    string `json:"args"`
}

func NewTask(name string, config SupervisorConfig) (*task, error) {
	t := &task{name: name, log: config.GetLog().(Log)}
	if err := json.Unmarshal(config.GetSettings(name), t); err != nil {
		return nil, err
	}
	if t.Command == "" {
		return nil, fmt.Errorf("Command for task %s is empty!", t.name)
	}
	if t.Args == "" {
		t.cmd = exec.Command(t.Command)
	} else {
		t.cmd = exec.Command(t.Command, strings.Split(t.Args, " ")...)
	}
	if err := t.setUser(); err != nil {
		return nil, err
	}
	t.setWorkDir()
	t.setEnv()
	return t, nil
}

func (t *task) Start() {
	Start(t)
}

func (t *task) setWorkDir() {
	if t.WorkDir != "" {
		t.cmd.Dir = t.WorkDir
	}
}

func (t *task) setEnv() {
	if t.Env == "" {
		t.cmd.Env = os.Environ()
	} else {
		t.cmd.Env = strings.Split(t.Env, " ")
	}
}

func (t *task) CommandWithArgs() string {
	return fmt.Sprintf("%s %s", t.Command, t.Args)
}
