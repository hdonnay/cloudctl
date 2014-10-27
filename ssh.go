package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/user"
	"sync"

	"code.google.com/p/go.crypto/ssh"
	"code.google.com/p/go.crypto/ssh/agent"
	"github.com/robertkrimen/otto"
)

type Ctx struct {
	sync.Mutex
	cfg     *ssh.ClientConfig
	clients map[string]*ssh.Client
}

func MakeCtx() *Ctx {
	sock, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		log.Println(err)
		return nil
	}
	ag := agent.NewClient(sock)
	name := *userArg
	if name == "" {
		u, err := user.Current()
		if err != nil {
			log.Println(err)
			return nil
		}
		name = u.Username
	}
	debug("agent: ", name)
	cfg := &ssh.ClientConfig{
		Auth:            []ssh.AuthMethod{ssh.PublicKeysCallback(ag.Signers)},
		User:            name,
		HostKeyCallback: nil,
	}
	debug("agent: ", cfg)
	return &Ctx{
		cfg:     cfg,
		clients: make(map[string]*ssh.Client),
	}
}

func (c *Ctx) Get(host string) (*ssh.Client, error) {
	c.Lock()
	defer c.Unlock()
	if cli, ok := c.clients[host]; ok {
		return cli, nil
	}
	var err error
	c.clients[host], err = ssh.Dial("tcp", net.JoinHostPort(host, "22"), c.cfg)
	if err != nil {
		debug("ctx.get:", err)
		return nil, err
	}
	return c.clients[host], nil
}

func (c *Ctx) Destroy(host string) error {
	c.Lock()
	defer c.Unlock()
	if cli, ok := c.clients[host]; ok {
		if err := cli.Close(); err != nil {
			return err
		}
		delete(c.clients, host)
	}
	return nil
}

func injectRun(ctx *Ctx, host string) func(otto.FunctionCall) otto.Value {
	return func(call otto.FunctionCall) otto.Value {
		client, err := ctx.Get(host)
		if err != nil {
			debug("run:", err)
			return otto.FalseValue()
		}
		s, err := client.NewSession()
		if err != nil {
			debug("run:", err)
			return otto.FalseValue()
		}
		o, err := s.StdoutPipe()
		if err != nil {
			debug("run:", err)
			return otto.FalseValue()
		}
		e, err := s.StderrPipe()
		if err != nil {
			debug("run:", err)
			return otto.FalseValue()
		}
		wg := &sync.WaitGroup{}
		wg.Add(3)
		go func() {
			defer wg.Done()
			r := bufio.NewReader(o)
			var line string
			var err error
			for line, err = r.ReadString('\n'); err != io.EOF; line, err = r.ReadString('\n') {
				fmt.Printf("[%s:%06s]\t%s", host, "stdout", line)
			}
			debug("output: ", err)
		}()
		go func() {
			defer wg.Done()
			r := bufio.NewReader(e)
			var line string
			var err error
			for line, err = r.ReadString('\n'); err != io.EOF; line, err = r.ReadString('\n') {
				fmt.Printf("[%s:%06s]\t%s", host, "stderr", line)
			}
			debug("output: ", err)
		}()
		go func() {
			defer wg.Done()
			in, err := s.StdinPipe()
			if err != nil {
				log.Println(err)
				return
			}
			defer in.Close()
			for i, a := range call.ArgumentList {
				debug("run: ", i, " ", a.String())
				fmt.Fprintln(in, a.String())
			}
		}()
		if err := s.Start("sh -s"); err != nil {
			log.Println(err)
			return otto.FalseValue()
		}
		defer s.Close()
		debug("run:", "running")
		s.Wait()
		wg.Wait()
		debug("run:", "ran")
		return otto.TrueValue()
	}
}
