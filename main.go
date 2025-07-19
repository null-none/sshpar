package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Script     string `yaml:"script"`
	Password   string `yaml:"password"`
	HostsFile  string `yaml:"hosts_file"`
	LogFile    string `yaml:"log_file"`
}

type HostConfig struct {
	Address  string
	Username string
	Port     string
}

func loadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("YAML parse error: %v", err)
	}
	return config
}

func parseHostLine(line string) HostConfig {
	defaultPort := "22"
	parts := strings.Split(line, "@")
	if len(parts) != 2 {
		log.Fatalf("Invalid host format: %s", line)
	}

	user := parts[0]
	hostPart := parts[1]

	host := hostPart
	port := defaultPort

	if strings.Contains(hostPart, ":") {
		hostParts := strings.Split(hostPart, ":")
		host = hostParts[0]
		port = hostParts[1]
	}

	return HostConfig{
		Address:  user + "@" + host,
		Username: user,
		Port:     port,
	}
}

func readHostsFile(filename string) []HostConfig {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening hosts file: %v", err)
	}
	defer file.Close()

	var hosts []HostConfig
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		hosts = append(hosts, parseHostLine(line))
	}

	return hosts
}

func readScript(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Cannot read script file %s: %v", path, err)
	}
	return string(content)
}

func sshExecute(host HostConfig, script string, password string, logWriter *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

	userHost := strings.Split(host.Address, "@")
	hostOnly := strings.Split(userHost[1], ":")[0]

	config := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	conn, err := ssh.Dial("tcp", hostOnly+":"+host.Port, config)
	if err != nil {
		msg := fmt.Sprintf("[%s] SSH connection error: %v\n", host.Address, err)
		logWriter.WriteString(msg)
		fmt.Print(msg)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		msg := fmt.Sprintf("[%s] SSH session error: %v\n", host.Address, err)
		logWriter.WriteString(msg)
		fmt.Print(msg)
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput(script)
	header := fmt.Sprintf("\n====== [%s] ======\n", host.Address)
	logWriter.WriteString(header)
	logWriter.WriteString(string(output))
	logWriter.WriteString("\n")

	fmt.Printf("%s%s\n", header, string(output))
	if err != nil {
		msg := fmt.Sprintf("[%s] Command execution error: %v\n", host.Address, err)
		logWriter.WriteString(msg)
		fmt.Print(msg)
	}
}

func main() {
	config := loadConfig("config.yaml")

	logFile, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	hosts := readHostsFile(config.HostsFile)
	script := readScript("templates/" + config.Script)

	var wg sync.WaitGroup
	for _, host := range hosts {
		wg.Add(1)
		go sshExecute(host, script, config.Password, logFile, &wg)
	}
	wg.Wait()

	fmt.Println("\n✅ All tasks completed. See:", config.LogFile)
}
