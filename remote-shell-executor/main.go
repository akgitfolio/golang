package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHClient represents an SSH client
type SSHClient struct {
	Host     string
	Port     int
	Username string
	Password string
	KeyFile  string
}

// NewSSHClient creates a new SSH client
func NewSSHClient(host string, port int, username, password, keyFile string) *SSHClient {
	return &SSHClient{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		KeyFile:  keyFile,
	}
}

// RunCommand runs a command on the remote server
func (client *SSHClient) RunCommand(command string) (string, error) {
	var auth []ssh.AuthMethod
	if client.Password != "" {
		auth = append(auth, ssh.Password(client.Password))
	} else if client.KeyFile != "" {
		key, err := ioutil.ReadFile(client.KeyFile)
		if err != nil {
			return "", err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return "", err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		return "", fmt.Errorf("no authentication method provided")
	}

	config := &ssh.ClientConfig{
		User:            client.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", client.Host, client.Port)
	connection, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return "", err
	}
	defer connection.Close()

	session, err := connection.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	if err := session.Run(command); err != nil {
		return stderrBuf.String(), err
	}

	return stdoutBuf.String(), nil
}

// ExecuteCommandsInParallel executes commands on multiple SSH clients in parallel
func ExecuteCommandsInParallel(clients []*SSHClient, command string) {
	var wg sync.WaitGroup
	results := make(chan string, len(clients))

	for _, client := range clients {
		wg.AddItem(client)
		go func(client *SSHClient) {
			defer wg.Done()
			output, err := client.RunCommand(command)
			if err != nil {
				results <- fmt.Sprintf("[%s] Error: %s\n", client.Host, err)
			} else {
				results <- fmt.Sprintf("[%s] Output: %s\n", client.Host, output)
			}
		}(client)
	}

	wg.Wait()
	close(results)

	for result := range results {
		fmt.Println(result)
	}
}

func main() {
	clients := []*SSHClient{
		NewSSHClient("192.168.1.100", 22, "user1", "password1", ""),
		NewSSHClient("192.168.1.101", 22, "user2", "password2", ""),
		NewSSHClient("192.168.1.102", 22, "user3", "", "/path/to/private/key"),
	}

	command := "uptime"
	ExecuteCommandsInParallel(clients, command)
}
