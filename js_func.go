package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robertkrimen/otto"
)

// This exports an "include" function into the javascript environment.
//
// When called, it effecitvely "eval"s the filename supplied as an argument.
// It returns 'false' if there was an error, and 'true' otherwise.
func JS_include(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()
	f, err := os.Open(filename)
	if err != nil {
		debug("include:", err)
		return otto.FalseValue()
	}
	s, err := call.Otto.Compile(f.Name(), f)
	if err != nil {
		debug("include:", err)
		return otto.FalseValue()
	}
	if _, err := call.Otto.Run(s); err != nil {
		debug("include:", err)
		return otto.FalseValue()
	}
	return otto.TrueValue()
}

/*
This exports a "shell" function into the javascript environment.

It takes an arguement that is a string to execute in a shell.

The string 'cloud.Shell' can be set to change the shell command.

The number 'cloud.ShellTimeout' can be tuned to kill the process after a
number of seconds. The "exit_code" of the returned object will be '-1' to
signify that the command was killed.

An exit_code of '-2' signifies an error in the Go function.

It returns an object that contains execution information:

	{
		"exit_code": 0,
		"stdout": "",
		"stderr": ""
	}
*/
func JS_shell(call otto.FunctionCall) otto.Value {
	vm := call.Otto
	ret, err := vm.ToValue(map[string]interface{}{
		"exit_code": int(-2),
		"stdout":    "",
		"stderr":    "",
	})
	if err != nil {
		debug("shell:", err)
		panic("whelp")
	}
	//ret.Object.Set("exit_code", -2)
	cfgRaw, err := vm.Get("cloud")
	if err != nil {
		debug("shell:", err)
		return ret
	}
	cfg := cfgRaw.Object()
	timeoutRaw, err := cfg.Get("ShellTimeout")
	if err != nil {
		debug("shell:", err)
		return ret
	}
	timeout, err := timeoutRaw.ToInteger()
	if err != nil {
		debug("shell:", err)
		return ret
	}
	shRaw, err := cfg.Get("Shell")
	if err != nil {
		debug("shell:", err)
		return ret
	}
	sh := strings.Fields(shRaw.String())
	if len(sh) == 0 {
		sh = []string{"sh"}
	}
	args := make([]string, len(call.ArgumentList))
	for i, v := range call.ArgumentList {
		args[i] = v.String()
	}
	sh = append(sh, "-c")
	sh = append(sh, strings.Join(args, " "))
	cmd := exec.Command(sh[0], sh[1:]...)
	cmd.Stdout = &bytes.Buffer{}
	cmd.Stderr = &bytes.Buffer{}
	done := make(chan struct{})
	go func() {
		if err := cmd.Run(); err != nil {
			ret.Object().Set("exit_code",
				err.(*exec.ExitError).Sys().(syscall.WaitStatus).ExitStatus())
		}
		close(done)
	}()
	defer func() {
		if recover() != nil {
			debug("there seems to be a race when using a 0 timeout, trying to get status or kill an exited process.")
		}
	}()
	if timeout == 0 {
		timeout++
	}
	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		if !cmd.ProcessState.Exited() {
			cmd.Process.Kill()
			// wait for the goroutine to mess with the state, so we can undo it...
			// may not be necesarry?
			<-done
		}
		ret.Object().Set("exit_code", -1)
	case <-done:
		ret.Object().Set("stdout", cmd.Stdout.(*bytes.Buffer).String())
		ret.Object().Set("stderr", cmd.Stderr.(*bytes.Buffer).String())
	}
	return ret
}

func JS_do(call otto.FunctionCall) otto.Value {
	cloudVal, err := call.Argument(0).Export()
	if err != nil {
		return otto.FalseValue()
	}
	cloud := cloudVal.(map[string]interface{})

	// create our worker pool
	wg := &sync.WaitGroup{}
	wg.Add(*pool)
	work := make(chan string)
	ctx := MakeCtx()
	for i := 0; i < *pool; i++ {
		go func() {
			defer wg.Done()
			for h := range work {
				// each host gets it own copy of the js runtime, now that it's populated.
				// we can then inject host-specific functions.
				vm := call.Otto.Copy()
				if err := vm.Set("_run", injectRun(ctx, h)); err != nil {
					log.Println(err)
					return
				}
				tasks, err := cloud["Tasks"].(otto.Value).Export()
				if err != nil {
					log.Println("unable to determine task list")
					return
				}
				for _, t := range tasks.([]interface{}) {
					debug("worker", i, t.(string))
					_, err := vm.Run("cloud._tasks." + t.(string) + "(_run)")
					if err != nil {
						log.Println(err)
						break
					}
				}
			}
		}()
	}

	hostVal, err := cloud["Hosts"].(otto.Value).Export()
	if err != nil {
		return otto.FalseValue()
	}
	hosts := make([]string, len(hostVal.([]interface{})))
	for i, s := range hostVal.([]interface{}) {
		hosts[i] = s.(string)
	}

	for _, h := range hosts {
		work <- h
	}
	close(work)
	wg.Wait()
	return otto.TrueValue()
}
