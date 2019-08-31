package tests

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type ExecFaker struct {
	testName string

	executionsMutex sync.Mutex
	executions      map[string]int
}

func NewExecFaker() *ExecFaker {
	return &ExecFaker{
		testName: callerFunctionName(),

		executionsMutex: sync.Mutex{},
		executions:      map[string]int{},
	}
}

func (e *ExecFaker) FakeExec(caseName string) func(name string, arg ...string) *exec.Cmd {
	return func(name string, arg ...string) *exec.Cmd {
		count := e.executions[caseName]

		cs := []string{"-test.run=" + e.testName, "--", "case-name=" + caseName, fmt.Sprintf("execution=%d", count), name}
		cs = append(cs, arg...)
		cmd := exec.Command(os.Args[0], cs...)

		e.executionsMutex.Lock()
		e.executions[caseName] = count + 1
		e.executionsMutex.Unlock()

		return cmd
	}
}

type ExecutionInfo struct {
	Args            []string
	Execution       int
	CaseName        string
	IsFakeExecution bool
}

func (e *ExecFaker) ExecInfo() ExecutionInfo {
	if len(os.Args) < 5 || os.Args[2] != "--" || !strings.HasPrefix(os.Args[3], "case-name=") || !strings.HasPrefix(os.Args[4], "execution=") {
		return ExecutionInfo{
			IsFakeExecution: false,
		}
	}

	executionStr := argValue(os.Args[4])
	execution, err := strconv.Atoi(executionStr)
	if err != nil {
		panic(err)
	}
	args := os.Args[5:]
	caseName := argValue(os.Args[3])

	return ExecutionInfo{
		Args:            args,
		Execution:       execution,
		CaseName:        caseName,
		IsFakeExecution: true,
	}
}

func FakeStdout(data []byte) []byte {
	str := string(data)
	idx := strings.Index(str, "PASS")
	if idx == -1 {
		return data
	}
	r := []byte(str[:idx])
	if len(r) == 0 {
		return nil
	}
	return r
}

func FakeStringStdout(str string) string {
	idx := strings.Index(str, "PASS")
	if idx == -1 {
		return str
	}
	r := str[:idx]
	if len(r) == 0 {
		return ""
	}
	return r
}

func callerFunctionName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fns := strings.Split(frame.Function, ".")
	return fns[len(fns)-1]
}

func argValue(arg string) string {
	args := strings.Split(arg, "=")
	if len(args) > 1 {
		return strings.Join(args[1:], "=")
	}
	return ""
}
