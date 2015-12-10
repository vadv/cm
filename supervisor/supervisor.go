package supervisor

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

const svRestartTimeout time.Duration = 1 * time.Second

type supervisor struct {
	mutex *sync.Mutex

	task      *task
	cmd       *exec.Cmd
	cmdStdErr chan string
	cmdStdOut chan string
	status    *syscall.WaitStatus

	log Log

	errChan chan error // внутренний канал с сообщениями об ошибках/выходе
}

func Start(task *task) {
	go start(task)
}

func start(task *task) {

	s := &supervisor{
		task:      task,
		errChan:   make(chan error),
		cmdStdErr: make(chan string),
		cmdStdOut: make(chan string),
		mutex:     &sync.Mutex{},
		log:       task.log,
	}

	go s.run()

	for {
		select {
		case err := <-s.errChan:
			if err != nil {
				s.log.Write("ERROR", "[%s] Task exited with error: %s\n", task.name, err.Error())
				s.updateStatus(err)
			} else {
				s.log.Write("ERROR", "[%s] Task exited with status code '0'\n", task.name)
			}
			s.restart()
		case str := <-s.cmdStdErr:
			s.log.Write("ERROR", "[%s] Task stderror: %#v\n", task.name, str)
		case str := <-s.cmdStdOut:
			s.log.Write("ERROR", "[%s] Task stdout: %#v\n", task.name, str)
		}
	}
}

func (s *supervisor) updateStatus(err error) {
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			s.status = &status
		}
	}
}

/* process states */

func (s *supervisor) processExited() bool {
	if s.status == nil {
		return true // exit 0
	}
	return s.status.Exited()
}

/*  process control */

func (s *supervisor) restart() {
	s.kill()
	time.Sleep(svRestartTimeout)
	go s.run()
}

func (s *supervisor) kill() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cmd == nil || s.cmd.Process == nil {
		return
	}
	if !s.processExited() {
		s.errChan <- s.cmd.Process.Kill()
	}
}

func (s *supervisor) run() {

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.log.Write("INFO", "[%s] Starting task: %#v\n", s.task.name, s.task.CommandWithArgs())

	s.cmd = &exec.Cmd{
		Path:        s.task.cmd.Path,
		Dir:         s.task.cmd.Dir,
		Args:        s.task.cmd.Args,
		Env:         s.task.cmd.Env,
		SysProcAttr: s.task.cmd.SysProcAttr,
	}

	s.feedOut()

	if err := s.cmd.Start(); err != nil {
		s.errChan <- err
		return
	}
	s.errChan <- s.cmd.Wait()
}

func (s *supervisor) feedOut() {
	stdout, _ := s.cmd.StdoutPipe()
	go feed(stdout, s.cmdStdOut)
	stderr, _ := s.cmd.StderrPipe()
	go feed(stderr, s.cmdStdErr)
}

func feed(pipe io.Reader, sink chan<- string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		sink <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		return
	}
}
