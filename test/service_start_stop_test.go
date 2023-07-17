package test

import (
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestServiceStartAndStop(t *testing.T) {
	makefilePath := ".."

	cmd := exec.Command("make", "start_service")
	cmd.Dir = makefilePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}

	time.Sleep(5 * time.Second)

	log.Println("Service is running, now trying to stop it")

	cmd = exec.Command("make", "stop_service")
	cmd.Dir = makefilePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to stop service: %v", err)
	}

	time.Sleep(5 * time.Second)

	log.Println("Service stopped successfully.")
}
