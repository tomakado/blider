package cmd

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

func Run(cmd *exec.Cmd) error {
	var output bytes.Buffer
	cmd.Stdout = &output

	var errs bytes.Buffer
	cmd.Stderr = &errs

	if err := cmd.Start(); err != nil {
		log.Fatalf("cmd.Start() error: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				log.Printf("Exit status: %d", status.ExitStatus())
			}
		} else {
			log.Fatalf("cmd.Wait() error: %v", err)
		}
	}

	outputStr := strings.TrimSpace(output.String())
	if len(outputStr) > 0 {
		log.Println(output.String())
	}

	errsStr := strings.TrimSpace(errs.String())
	if len(errsStr) > 0 {
		log.Println(errs.String())
	}

	return nil
}
