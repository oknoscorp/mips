package command

import (
	"bufio"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type Command struct {
	IP      string
	Command string
}

// New will initialize command component.
func New(cmd Command) *Command {
	return &cmd
}

// Execute will send TCP request to desired IP with desired
// package content.
func (cmd *Command) Execute() (string, error) {

	dialer := net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.Dial("tcp", cmd.IP+":4028")
	if err != nil {
		log.Error(err)
		return "Dialing failed", err
	}

	defer conn.Close()

	conn.Write([]byte(cmd.Command))
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	response, err := bufio.NewReader(conn).ReadString('\x00')
	if err != nil {
		log.Error(err)
		return "", err
	}

	return strings.TrimRight(response, "\x00"), nil
}
