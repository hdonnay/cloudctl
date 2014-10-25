package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/robertkrimen/otto"
)

var (
	ctljs    = flag.String("main", "ctl.js", "main control file")
	rolesArg = flag.String("R", "", "roles to execute task on")
)

func main() {
	flag.Parse()
	vm := otto.New()
	cfg, _ := vm.Object("cloud = {};")
	cfg.Set("Shell", "sh -c")
	cfg.Set("ShellTimeout", 10)
	cfg.Set("Roles", strings.Split(*rolesArg, ","))

	if err := prepVM(vm, cfg); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	f, err := os.Open(*ctljs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	m, err := vm.Compile("", f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if _, err := vm.Run(m); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func prepVM(vm *otto.Otto, cfg *otto.Object) error {
	if err := vm.Set("cloud", cfg); err != nil {
		return err
	}
	debug("cloud object injected")
	if err := vm.Set("include", JS_include(vm)); err != nil {
		return err
	}
	debug("include function injected")
	if err := vm.Set("shell", JS_shell(vm)); err != nil {
		return err
	}
	debug("shell function injected")
	return nil
}
