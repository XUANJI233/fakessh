package main

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"bufio"
	"io"

	"golang.org/x/crypto/ssh"
)

const (
	logPath = "/var/log/fakessh.log"
	errLogPath = "/var/log/wrong.log"
	maxLogEntries = 1000
)

var (
	errBadPassword = errors.New("permission denied")
	serverVersions = []string{
		"SSH-2.0-OpenSSH_6.6.1p1 Ubuntu-2ubuntu2.3",
		"SSH-2.0-OpenSSH_6.7p1 Debian-5+deb8u3",
		"SSH-2.0-OpenSSH_7.2p2 Ubuntu-4ubuntu2.10",
		"SSH-2.0-OpenSSH_7.4",
		"SSH-2.0-OpenSSH_8.0",
		"SSH-2.0-OpenSSH_8.4p1 Debian-2~bpo10+1",
		"SSH-2.0-OpenSSH_8.4p1 Debian-5+deb11u1",
		"SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6",
	}
	logMutex sync.Mutex // A mutex to protect concurrent writes to the log file.
	errLogger *log.Logger // Logger for error messages
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	errLogFile, err := os.OpenFile(errLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open the error log file: %v", err)
		return
	}
	defer errLogFile.Close()

	errLogger = log.New(errLogFile, "", log.LstdFlags | log.Lmicroseconds)

	serverConfig := &ssh.ServerConfig{
		MaxAuthTries:     6,
		PasswordCallback: passwordCallback,
		ServerVersion:    serverVersions[0],
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		errLogger.Printf("Failed to generate private key: %v", err) // Use the error logger here.
		return
	}

	signer, err := ssh.NewSignerFromSigner(privateKey)
	if err != nil {
		errLogger.Printf("Failed to create signer: %v", err) // And here.
		return
	}

	serverConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", ":22")
	if err != nil {
		errLogger.Println("Failed to listen:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			errLogger.Println("Failed to accept:", err)
			break
		}
		go handleConn(conn, serverConfig)
	}
}

func passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	entry := fmt.Sprintf("%s %s %s %s\n", conn.RemoteAddr(), string(conn.ClientVersion()), conn.User(), string(password))
	
	lines, err := readLastLines(logPath, maxLogEntries - 1)
	if err != nil && !os.IsNotExist(err) {
		errLogger.Println("Failed to read log file:", err)
		return nil, errBadPassword
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	file, err := os.Create(logPath)
	if err != nil {
		errLogger.Println("Failed to open log file:", err)
		return nil, errBadPassword
	}
	defer file.Close()

	for _, line := range lines {
		_, _ = file.WriteString(line + "\n")
	}
	_, _ = file.WriteString(entry)

	time.Sleep(100 * time.Millisecond)
	return nil, errBadPassword
}

func readLastLines(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0, n)

	for scanner.Scan() {
		line := scanner.Text()
		if len(lines) < n {
			lines = append(lines, line)
		} else {
			copy(lines, lines[1:])
			lines[n-1] = line
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return lines, nil
}

func handleConn(conn net.Conn, serverConfig *ssh.ServerConfig) {
	defer conn.Close()
	log.Println(conn.RemoteAddr())
	ssh.NewServerConn(conn, serverConfig)
}
