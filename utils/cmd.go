package utils

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func ExecuteCmd(c string, dir string, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	out, err := cmd.Output()
	if err != nil {
		return string(out[:]) + err.Error(), true
	}

	if string(out[:]) == "" {
		return "Command executed successfully but no output", false
	}

	return string(out[:]), false
}

func RunBlockingWithTerminalIO(command string, dir string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		if cmd.Process != nil {
			_ = cmd.Process.Signal(sig)
		}
	}()

	err := cmd.Run()

	signal.Stop(sigs)
	close(sigs)

	return err
}
