package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/flike/golog"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:6380")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}
		handleConn(conn)
	}
}

func handleConn(conn net.Conn) error {
	for {
		request, err := newRequest(conn)
		if err != nil {
			return err
		}

		reply := serveRequest(request)
		if _, err := reply.WriteTo(conn); err != nil {
			golog.Error("server", "onConn", "reply write error", 0,
				"err", err.Error())
			return err
		}
	}
}
func newRequest(conn io.ReadCloser) (*Request, error) {
	connReader := bufio.NewReader(conn)
	line, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var argCount int
	if line[0] == '*' {
		if _, err := fmt.Sscanf(line, "*%d\r\n", &argCount); err != nil {
			fmt.Println(argCount)
			return nil, err
		}
		command, err := readArgument(connReader)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		arguments := make([][]byte, argCount-1)
		for i := 0; i < argCount-1; i++ {
			if arguments[i], err = readArgument(connReader); err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
		request := &Request{}
		request.Command = string(command)
		request.Arguments = arguments
		request.Conn = conn
		return request, nil
	}
	return nil, errors.New("coming message error format")
}

func serveRequest(request *Request) Reply {
	switch request.Command {
	case "GET":
		return nil
	case "SET":
		return nil
	case "EXISTS":
		return nil
	case "DEL":
		return nil
	case "SELECT":
		return nil
	default:
		return nil
	}
}

type Request struct {
	Command   string
	Arguments [][]byte
	Conn      io.ReadCloser
}

func readArgument(reader *bufio.Reader) ([]byte, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var argLength int
	if _, err := fmt.Sscanf(line, "$%d\r\n", &argLength); err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(io.LimitReader(reader, int64(argLength)))
	if err != nil {
		return nil, err
	}
	if b, err := reader.ReadByte(); err != nil || b != '\r' {
		return nil, errors.New("\r not found")
	}
	if b, err := reader.ReadByte(); err != nil || b != '\n' {
		return nil, errors.New("\n not found")
	}
	return data, nil
}

type Reply io.WriterTo

type ErrorReply struct {
	message string
}

func (reply *ErrorReply) WriteTo(writer io.Writer) (int64, error) {
	n, err := writer.Write([]byte("-ERROR " + reply.message + "\r\n"))
	return int64(n), err
}
func (reply *ErrorReply) Error() string {
	return reply.message
}

type IntReply struct {
	number int64
}

func (reply IntReply) WriteTo(writer io.Writer) (int64, error) {
	n, err := writer.Write([]byte(":" + strconv.FormatInt(reply.number, 10) + "\r\n"))
	return int64(n), err
}

type StatusReply struct {
	code string
}

func (reply StatusReply) WriteTo(writer io.Writer) (int64, error) {
	n, err := writer.Write([]byte("+" + reply.code + "\r\n"))
	return int64(n), err
}

type BulkReply struct {
	value []byte
}

func (reply BulkReply) WriterTo(writer io.Writer) (int64, error) {
	n, err := writeBytes(reply.value, writer)
	return 0, nil
}

type MultiBulkReply struct {
	values [][]byte
}

func (reply *MultiBulkReply) WriterTo(writer io.Writer) (int64, error) {
	if reply.values == nil {
		return 0, errors.New("value nil")
	}
	n := 0
	if n, err := writer.Write([]byte("*" + strconv.Itoa(len(reply.values)) + "\r\n")); err != nil {
		return int64(n), err
	}
	for _, v := range reply.values {
		wrote, err := writeBytes(v, writer)
		n += int(wrote)
		if err != nil {
			return int64(n), err
		}
	}
	return int64(n), nil
}

func writeNullByte(writer io.Writer) (int64, error) {
	n, err := writer.Write([]byte("$-1\r\n"))
	return int64(n), err
}

func writeBytes(value interface{}, writer io.Writer) (int64, error) {
	if value == nil {
		return writeNullByte(writer)
	}
	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			return writeNullByte(writer)
		}
		buf := []byte("$" + strconv.Itoa(len(v)) + "\r\n")
		buf = append(buf, v...)
		buf = append(buf, []byte("\r\n")...)
		n, err := writer.Write(buf)
		return int64(n), err
	case int:
		n, err := writer.Write([]byte(":" + strconv.Itoa(v) + "\r\n"))
		return int64(n), err
	}
}
