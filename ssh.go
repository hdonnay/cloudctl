package main

import (
	"net"

	"code.google.com/p/go.crypto/ssh"
	"github.com/robertkrimen/otto"
)

type Ctx struct {
	client *ssh.Client
}

func injectRemote(vm *otto.Otto, sshcfg *ssh.ClientConfig, m *map[string]*Ctx) error {
	conn := *m
	return vm.Set("Remote", func(call otto.FunctionCall) otto.Value {
		host := call.Argument(0).String()
		debug("Remote:", host)

		if _, ok := conn[host]; !ok {
			debug("Remote:", host, "making ssh connection")
			c, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), sshcfg)
			if err != nil {
				panic(err)
			}
			conn[host] = &Ctx{
				client: c,
			}
		}
		ret, err := call.Otto.Object(`{}`)
		if err != nil {
			debug("Remote:", err)
			return otto.FalseValue()
		}
		ret.Set("run", func(call otto.FunctionCall) otto.Value {
			debug("Remote.run:", host)
			fn, err := call.Argument(0).Call(call.Argument(0))
			if err != nil {
				debug("Remote.run:", host, err)
			}
			return fn
		})
		return ret.Value()
	})
}
