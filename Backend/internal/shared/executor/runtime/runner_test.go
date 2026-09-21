package runtime

import (
	"context"
	"os"
	"testing"
)

func TestRunCapturesOutputAndExitCode(t *testing.T) {
	result, err := runHelper(context.Background(), "output")
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if result.Output != "helper output\n" || result.ExitCode != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestRunReturnsNonZeroExitCode(t *testing.T) {
	result, err := runHelper(context.Background(), "fail")
	if err == nil || result.ExitCode != 7 {
		t.Fatalf("unexpected result=%+v, err=%v", result, err)
	}
}

func TestRunReportsStartFailure(t *testing.T) {
	result, err := Run(context.Background(), "missing-executor-program", nil, "ignored")
	if err == nil || result.ExitCode != -1 {
		t.Fatalf("unexpected result=%+v, err=%v", result, err)
	}
}

func runHelper(ctx context.Context, mode string) (result struct {
	Output   string
	ExitCode int
}, err error) {
	runResult, runErr := Run(ctx, os.Args[0], []string{"-test.run=TestHelperProcess", "--"}, mode)
	return struct {
		Output   string
		ExitCode int
	}{Output: runResult.Output, ExitCode: runResult.ExitCode}, runErr
}

func TestHelperProcess(t *testing.T) {
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "output":
		_, _ = os.Stdout.WriteString("helper output\n")
		os.Exit(0)
	case "fail":
		_, _ = os.Stdout.WriteString("helper output\n")
		os.Exit(7)
	}
}
