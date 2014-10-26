package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

var (
	ctljs    = flag.String("main", "ctl.js", "main control file")
	rolesArg = flag.String("R", "", "roles to execute task on")
	hostsArg = flag.String("H", "", "hosts to execute task on")
	pool     = flag.Int("z", 5, "hosts to run on at one time")
)

func main() {
	sockPath := os.Getenv("SSH_AUTH_SOCK")
	sock, err := net.Dial("unix", sockPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ag := agent.NewClient(sock)
	flag.Parse()
	ctx := make(map[string]*Ctx)
	vm := otto.New()
	sshCfg := &ssh.ClientConfig{
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(ag.Signers),
		},
	}
	cfg, _ := vm.Object("cloud = {};")
	cfg.Set("Shell", "sh")
	cfg.Set("ShellTimeout", 10)
	cfg.Set("Roles", strings.Split(*rolesArg, ","))
	cfg.Set("Hosts", strings.Split(*hostsArg, ","))
	cfg.Set("Tasks", flag.Args())

	if err := prepVM(vm, cfg, &ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := injectRemote(vm, sshCfg, &ctx); err != nil {
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
	if _, err := vm.Run(mainJS); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func prepVM(vm *otto.Otto, cfg *otto.Object, ctx *map[string]*Ctx) error {
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
	if err := vm.Set("run", injectRun(ctx)); err != nil {
		return err
	}
	if err := vm.Set("_do", JS_do); err != nil {
		return err
	}
	debug("_do function injected")
	return nil
}
