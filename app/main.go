package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	cmd := os.Args[3]
	args := os.Args[4:len(os.Args)]

	if err := run(cmd, args); err != nil {
		log.Fatal(err)
	}
}

func run(name string, args []string) error {
	rootDir := filepath.Join(os.TempDir(), "mydocker", uitoa(uint(rand.Uint32())))
	if err := os.MkdirAll(filepath.Join(rootDir, filepath.Dir(name)), os.ModeDir); err != nil {
		return fmt.Errorf("failed to create rootdir: %w", err)
	}
	defer os.RemoveAll(rootDir)

	src, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	srcInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file status: %w", err)
	}

	dst, err := os.OpenFile(filepath.Join(rootDir, name), os.O_CREATE|os.O_WRONLY, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to open destination file: %w", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file from %s to %s: %w", src.Name(), dst.Name(), err)
	}

	if err := src.Close(); err != nil {
		return fmt.Errorf("failed to close source file: %w", err)
	}
	if err := dst.Close(); err != nil {
		return fmt.Errorf("failed to close destination file: %w", err)
	}

	// workaround for chroot
	if err := os.Mkdir(filepath.Join(rootDir, "dev"), os.ModeDir); err != nil {
		if !os.IsExist(err) {
			return fmt.Errorf("failed to create /dev directory: %w", err)
		}
		devnull, err := os.Create(filepath.Join(rootDir, "/dev/null"))
		if err != nil {
			return fmt.Errorf("failed to create /dev/null file: %w", err)
		}
		if err := devnull.Close(); err != nil {
			return fmt.Errorf("failed to close /dev/null file: %w", err)
		}
	}

	chrootArgs := []string{rootDir, name}
	chrootArgs = append(chrootArgs, args...)
	cmd := exec.Command("chroot", chrootArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("fail to run command: %w", err)
	}
	return nil
}

func uitoa(val uint) string {
	if val == 0 { // avoid string allocation
		return "0"
	}
	var buf [20]byte // big enough for 64bit value base 10
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	// val < 10
	buf[i] = byte('0' + val)
	return string(buf[i:])
}
