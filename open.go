package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenAsNeeded(doit bool, path string) {
	if !doit {
		return
	}
	if openCmd := getOpenCmd(path); openCmd != nil {
		err := openCmd.Start()
		if err != nil {
			log.Printf("Failed to open %s: %s", path, err)
		}
		return
	}
	log.Printf("Don't know how to open output on this platform.")
}

func getOpenCmd(path string) *exec.Cmd {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path)
	case "linux":
		return exec.Command("xdg-open", path)
	case "windows":
		cmd := "url.dll,FileProtocolHandler"
		runDll32 := filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
		return exec.Command(runDll32, cmd, path)
	default:
		return nil
	}
}
