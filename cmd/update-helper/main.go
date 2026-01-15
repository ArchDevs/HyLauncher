package main

import (
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		return
	}

	oldExe := os.Args[1]
	newExe := os.Args[2]

	time.Sleep(2 * time.Second)

	_ = os.Remove(oldExe)
	_ = os.Rename(newExe, oldExe)

	_ = exec.Command(oldExe).Start()
}
