package zfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestRunPipeDestinationExitsEarly(t *testing.T) {
	for _, code := range []int{0, 23} {
		t.Run(strconv.Itoa(code), func(t *testing.T) {
			ctx := runPipeTestContext(t)
			src := runPipeTestCommand(ctx, "write")
			dst := runPipeTestCommand(ctx, "exit", strconv.Itoa(code))
			if _, err := runPipe(src, dst, "test"); err == nil {
				t.Fatal("an incomplete transfer must fail")
			}
			if src.ProcessState == nil || dst.ProcessState == nil {
				t.Fatal("both child processes must have been waited for")
			}
		})
	}
}

func TestRunPipeReportsSourceFailure(t *testing.T) {
	for _, destination := range []string{"copy", "idle"} {
		t.Run(destination, func(t *testing.T) {
			ctx := runPipeTestContext(t)
			src := runPipeTestCommand(ctx, "exit", "7")
			dst := runPipeTestCommand(ctx, destination)
			out, err := runPipe(src, dst, "test")
			requirePipeExitCode(t, err, 7)
			if !bytes.Contains(out, []byte("fixture failure: 7")) {
				t.Fatalf("source diagnostic missing from %q", out)
			}
			if src.ProcessState == nil || dst.ProcessState == nil {
				t.Fatal("both child processes must have been waited for")
			}
		})
	}
}

func TestRunPipeReportsDestinationFailure(t *testing.T) {
	ctx := runPipeTestContext(t)
	payload := []byte("complete stream\n")
	src := runPipeTestCommand(ctx, "copy")
	src.Stdin = bytes.NewReader(payload)
	dst := runPipeTestCommand(ctx, "copy-fail")
	out, err := runPipe(src, dst, "test")
	requirePipeExitCode(t, err, 23)
	if src.ProcessState == nil || !src.ProcessState.Success() {
		t.Fatal("source must complete successfully before destination failure")
	}
	if !bytes.HasPrefix(out, payload) || !bytes.Contains(out, []byte("fixture failure: 23")) {
		t.Fatalf("destination output or diagnostic missing from %q", out)
	}
	if dst.ProcessState == nil {
		t.Fatal("the destination process must have been waited for")
	}
}

func TestRunPipeDestinationFailureStopsIdleSource(t *testing.T) {
	ctx := runPipeTestContext(t)
	src := runPipeTestCommand(ctx, "idle")
	dst := runPipeTestCommand(ctx, "exit", "23")
	out, err := runPipe(src, dst, "test")
	requirePipeExitCode(t, err, 23)
	if !bytes.Contains(out, []byte("fixture failure: 23")) {
		t.Fatalf("destination diagnostic missing from %q", out)
	}
	if src.ProcessState == nil || dst.ProcessState == nil {
		t.Fatal("both child processes must have been waited for")
	}
}

func TestRunPipeTransfersDataAndFinalizes(t *testing.T) {
	ctx := runPipeTestContext(t)
	payload := bytes.Repeat([]byte("zfs stream\n"), 128*1024)
	src := runPipeTestCommand(ctx, "copy")
	src.Stdin = bytes.NewReader(payload)
	dst := runPipeTestCommand(ctx, "finalize")
	out, err := runPipe(src, dst, "test")
	if err != nil {
		t.Fatal(err)
	}
	want := append(payload, []byte("finalized\n")...)
	if !bytes.Equal(out, want) {
		t.Fatalf("transfer mismatch: got %d bytes, want %d", len(out), len(want))
	}
	if src.ProcessState == nil || dst.ProcessState == nil {
		t.Fatal("both child processes must have been waited for")
	}
}

func TestRunPipeSourceStartFailure(t *testing.T) {
	ctx := runPipeTestContext(t)
	src := exec.CommandContext(ctx, filepath.Join(t.TempDir(), "missing"))
	dst := runPipeTestCommand(ctx, "copy")
	if _, err := runPipe(src, dst, "test"); err == nil {
		t.Fatal("source start failure must be reported")
	}
	if dst.Process != nil {
		t.Fatal("destination must not start when source start fails")
	}
}

func TestRunPipeDestinationStartFailureReapsSource(t *testing.T) {
	ctx := runPipeTestContext(t)
	src := runPipeTestCommand(ctx, "idle")
	dst := exec.CommandContext(ctx, filepath.Join(t.TempDir(), "missing"))
	if _, err := runPipe(src, dst, "test"); err == nil {
		t.Fatal("destination start failure must be reported")
	}
	if src.ProcessState == nil {
		t.Fatal("the source process must have been waited for when the destination fails to start")
	}
}

// Re-execute the test binary for real process/pipe behavior without external tools.
func TestRunPipeProcess(t *testing.T) {
	action := os.Getenv("RUNPIPE_TEST_PROCESS")
	if action == "" {
		return
	}
	switch action {
	case "copy", "finalize", "copy-fail":
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			os.Exit(1)
		}
		if action == "copy-fail" {
			fmt.Fprint(os.Stderr, "fixture failure: 23\n")
			os.Exit(23)
		}
		if action == "finalize" {
			// The consumer must be allowed to finish work after producer EOF.
			time.Sleep(time.Second)
			fmt.Fprint(os.Stdout, "finalized\n")
		}
	case "write":
		chunk := bytes.Repeat([]byte("zfs"), 16*1024)
		for {
			if _, err := os.Stdout.Write(chunk); err != nil {
				os.Exit(1)
			}
		}
	case "exit":
		code, err := strconv.Atoi(os.Args[len(os.Args)-1])
		if err != nil {
			t.Fatal(err)
		}
		if code != 0 {
			fmt.Fprintf(os.Stderr, "fixture failure: %d\n", code)
		}
		os.Exit(code)
	case "idle":
		time.Sleep(time.Hour)
	default:
		t.Fatalf("unknown process fixture %q", action)
	}
	os.Exit(0)
}

func runPipeTestContext(t *testing.T) context.Context {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(func() {
		if ctx.Err() != nil {
			t.Error("runPipe hung until the watchdog killed its children")
		}
		cancel()
	})
	return ctx
}

func runPipeTestCommand(ctx context.Context, action string, args ...string) *exec.Cmd {
	args = append([]string{"-test.run=^TestRunPipeProcess$", "--"}, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], args...)
	cmd.Env = append(os.Environ(), "RUNPIPE_TEST_PROCESS="+action)
	return cmd
}

func requirePipeExitCode(t *testing.T, err error, want int) {
	t.Helper()

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != want {
		t.Fatalf("want first exit code %d, got %v", want, err)
	}
}
