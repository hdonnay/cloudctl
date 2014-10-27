package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

var (
	ctljs      = flag.String("main", "ctl.js", "main control file")
	rolesArg   = flag.String("R", "", "roles to execute task on")
	hostsArg   = flag.String("H", "", "hosts to execute task on")
	pool       = flag.Int("z", 5, "hosts to run on at one time")
	userArg    = flag.String("u", "", "username to use")
	restAsTask = flag.Bool("c", false, "use non-flag arguments as a run() task")
)

func main() {
	flag.Parse()
	vm := otto.New()
	if err := vm.Set("cloud", make(map[string]interface{})); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfgVal, err := vm.Get("cloud")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg := cfgVal.Object()
	if cfg == nil {
		fmt.Println("config is nil")
		os.Exit(1)
	}
	cfg.Set("Shell", "sh")
	cfg.Set("ShellTimeout", 10)
	cfg.Set("Roles", strings.Split(*rolesArg, ","))
	cfg.Set("Hosts", strings.Split(*hostsArg, ","))
	cfg.Set("Tasks", flag.Args())
	cfg.Set("_tasks", make(map[string]interface{}))

	if err := prepVM(vm, cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	f, err := os.Open(*ctljs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m, err := vm.Compile(f.Name(), f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := vm.Run(m); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *restAsTask {
		// magic up a task
		if _, err := vm.Run(fmt.Sprintf(restTask, strings.Join(flag.Args(), " "))); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	if _, err := vm.Run(mainJS); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func prepVM(vm *otto.Otto, cfg *otto.Object) error {
	if _, err := vm.Run(env); err != nil {
		return err
	}
	debug("environment injected")
	if err := vm.Set("cloud", cfg); err != nil {
		return err
	}
	debug("cloud object injected")
	if err := vm.Set("include", JS_include); err != nil {
		return err
	}
	debug("include function injected")
	if err := vm.Set("shell", JS_shell); err != nil {
		return err
	}
	debug("shell function injected")
	if err := vm.Set("_do", JS_do); err != nil {
		return err
	}
	debug("_do function injected")
	return nil
}
