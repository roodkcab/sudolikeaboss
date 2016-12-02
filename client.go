package main

import (
	"fmt"
	"github.com/roodkcab/sudolikeaboss/onepass"
	"os"
	"strconv"
	"time"
    "os/exec"
    "strings"
)

const DEFAULT_TIMEOUT_STRING_SECONDS = "30"
const DEFAULT_HOST = "chrome://extension"
const DEFAULT_WEBSOCKET_URI = "ws://127.0.0.1:6263/4"
const DEFAULT_WEBSOCKET_PROTOCOL = ""
const DEFAULT_WEBSOCKET_ORIGIN = "chrome-extension://aomjjhallfgjeglblehebfpbcfeobpgk"

func LoadConfiguration() *onepass.Configuration {
	defaultHost := os.Getenv("SUDOLIKEABOSS_DEFAULT_HOST")
	if defaultHost == "" {
		defaultHost = DEFAULT_HOST
	}

	websocketUri := os.Getenv("SUDOLIKEABOSS_WEBSOCKET_URI")
	if websocketUri == "" {
		websocketUri = DEFAULT_WEBSOCKET_URI
	}

	websocketProtocol := os.Getenv("SUDOLIKEABOSS_WEBSOCKET_PROTOCOL")
	if websocketProtocol == "" {
		websocketProtocol = DEFAULT_WEBSOCKET_PROTOCOL
	}

	websocketOrigin := os.Getenv("SUDOLIKEABOSS_WEBSOCKET_ORIGIN")
	if websocketOrigin == "" {
		websocketOrigin = DEFAULT_WEBSOCKET_ORIGIN
	}

	return &onepass.Configuration{
		WebsocketUri:      websocketUri,
		WebsocketProtocol: websocketProtocol,
		WebsocketOrigin:   websocketOrigin,
		DefaultHost:       defaultHost,
	}
}

func retrievePasswordFromOnepassword(configuration *onepass.Configuration, done chan bool) {
	// Load configuration from a file
	client, err := onepass.NewClientWithConfig(configuration)

	if err != nil {
		os.Exit(1)
	}

	authResponse, err := client.SendHelloCommand()

	if err != nil {
		fmt.Println(authResponse)
		os.Exit(1)
	}

	response, err := client.SendShowPopupCommand()

	if err != nil {
		fmt.Println(response)
		os.Exit(1)
	}

    host := strings.Split(response.Url, " ")

    base := append([]string{host[0], response.Script[1][2]}, host[1:]...)
    //sshArgs := append(base)
    cmd         := exec.Command("/opt/local/bin/ssh.sh", base...)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Run()

	/*password, err := response.GetPassword()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(password)*/

	done <- true
}

// Run the main sudolikeaboss entry point
func runSudolikeaboss() {
	done := make(chan bool)

	configuration := LoadConfiguration()

	timeoutString := os.Getenv("SUDOLIKEABOSS_TIMEOUT_SECS")
	if timeoutString == "" {
		timeoutString = DEFAULT_TIMEOUT_STRING_SECONDS
	}

	timeout, err := strconv.ParseInt(timeoutString, 10, 16)

	if err != nil {
		os.Exit(1)
	}

	go retrievePasswordFromOnepassword(configuration, done)

	// Timeout if necessary
	select {
	case <-done:
		// Do nothing no need
	case <-time.After(time.Duration(timeout) * time.Second):
		close(done)
		os.Exit(1)
	}
	// Close the app neatly
	os.Exit(0)
}
