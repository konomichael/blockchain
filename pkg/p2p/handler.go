package p2p

import (
	"bufio"

	"github.com/libp2p/go-libp2p/core/network"
)

const commandLength = 12

type Handler func([]byte, *bufio.ReadWriter)

func MakeStreamHandler(handlers map[string]Handler) network.StreamHandler {
	return func(s network.Stream) {
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		msg, err := rw.ReadBytes('\n')
		if err != nil {
			return
		}
		cmd, data := readCommand(string(msg))
		if handler, ok := handlers[cmd]; ok {
			handler(data, rw)
		}
	}
}

func readCommand(msg string) (string, []byte) {
	if len(msg) < commandLength+1 {
		return "", nil
	}
	cmdBytes, dataBytes := msg[:commandLength], msg[commandLength:len(msg)-1]
	var cmd []byte
	for i := range cmdBytes {
		if cmdBytes[i] != 0x00 {
			cmd = append(cmd, cmdBytes[i])
		}
	}

	return string(cmd), []byte(dataBytes)
}

func SendCommand(rw *bufio.ReadWriter, cmd string, data []byte) error {
	msg := make([]byte, commandLength)
	copy(msg, cmd)
	msg = append(msg, data...)
	msg = append(msg, '\n')
	_, err := rw.Write(msg)
	if err != nil {
		return err
	}
	return rw.Flush()
}
