// Copyright 2014 Paul Hammond.
// This software is licensed under the MIT license, see LICENSE.txt for details.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"strings"

	"github.com/ogier/pflag"
)

// Config defines structure of the config file.
type Config struct {
	HipchatURL string `json:"hipchat_url"`
	Room       string `json:"room"`
	APIToken   string `json:"api_token"`
}

// Load loads in a found config file.
func (c *Config) Load() error {
	err := c.loadConfigFiles()
	if err != nil {
		return err
	}
	c.loadEnvVars()

	if c.HipchatURL == "" {
		return errors.New("Could not find a HipchatURL in HIPCHAT_URL, /etc/hipcat.conf, /.hipcat.conf, ./hipcat.conf")
	}

	if c.APIToken == "" {
		return errors.New("Could not find an APIToken in HIPCAT_API_TOKEN, /etc/hipcat.conf, /.hipcat.conf, ./hipcat.conf")
	}

	return nil
}

func (c *Config) loadEnvVars() {
	envs := []string{"HIPCHAT_URL", "HIPCAT_ROOM", "HIPCAT_API_TOKEN"}
	for _, env := range envs {
		envVal := os.Getenv(env)
		if envVal == "" {
			continue
		}

		switch env {
		case "HIPCHAT_URL":
			c.HipchatURL = envVal
		case "HIPCAT_ROOM":
			c.Room = envVal
		case "HIPCAT_API_TOKEN":
			c.APIToken = envVal
		}
	}
}

func (c *Config) loadConfigFiles() error {
	homeDir := ""
	usr, err := user.Current()
	if err == nil {
		homeDir = usr.HomeDir
	}

	for _, path := range []string{"/etc/hipcat.conf", homeDir + "/.hipcat.conf", "./hipcat.conf"} {
		file, err := os.Open(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}

		err = json.NewDecoder(file).Decode(c)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) bindFlags() {
	pflag.StringVarP(&c.Room, "room", "r", c.Room, "room")
}

// RoomMessage contains the message for the Hipchat room.
type RoomMessage struct {
	Message string `json:"message"`
}

// Encode encodes the RoomMessage into a json string.
func (m RoomMessage) Encode() (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Post posts the RoomMessage to the Hipchat server.
func (m RoomMessage) Post(config *Config) error {
	encoded, err := m.Encode()
	if err != nil {
		return err
	}

	URL, err := url.Parse(fmt.Sprintf("%s/v2/room/%s/message",
		config.HipchatURL, config.Room))
	if err != nil {
		return err
	}

	var req *http.Request
	req, err = http.NewRequest("POST", URL.String(), bytes.NewBuffer([]byte(encoded)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("Not OK: %d, %s", resp.StatusCode, body)
	}
	return nil
}

func main() {
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: hipcat [-r room] [message]")
	}

	cfg := Config{}
	err := cfg.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	cfg.bindFlags()
	pflag.Parse()

	if cfg.Room == "" {
		log.Println("Could not find a Room in HIPCAT_ROOM, /etc/hipcat.conf, /.hipcat.conf, ./hipcat.conf or passed with -r")
		pflag.Usage()
		os.Exit(1)
	}
	// was there a message on the command line? If so use it.
	args := pflag.Args()
	if len(args) > 0 {
		msg := RoomMessage{
			Message: strings.Join(args, " "),
		}

		err = msg.Post(&cfg)
		if err != nil {
			log.Fatalf("Post failed: %v", err)
		}
		return
	}

	// ...Otherwise scan stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := RoomMessage{
			Message: scanner.Text(),
		}

		err = msg.Post(&cfg)
		if err != nil {
			log.Fatalf("Post failed: %v", err)
		}
	}
	if err = scanner.Err(); err != nil {
		log.Fatalf("Error reading: %v", err)
	}
}
