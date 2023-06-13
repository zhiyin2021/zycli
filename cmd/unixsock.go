package cmd

import (
	"bufio"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiyin2021/zycli/tools"
)

var (
	IPCMsg    func(msg string) string
	sock      = tools.CurrentName() + ".ipc"
	ulistener net.Listener
)

func stopUnixSock() {
	if ulistener != nil {
		os.Remove(sock)
	}
}
func startUnixSock() error {
	// addr, _ := net.ResolveUnixAddr("unix", sock)
	var err error
	ulistener, err = net.Listen("unix", sock)
	if err != nil {
		return err
	}
	go func() {
		defer ulistener.Close()
		for {
			conn, err := ulistener.Accept()
			if err != nil {
				continue
			}
			go func(con net.Conn) {
				defer con.Close()
				reader := bufio.NewReader(conn)
				msg, _, err := reader.ReadLine()
				if err != nil {
					return
				}
				logrus.Println("uread:", msg)
				message := string(msg)
				if message == "dbg" {
					DEBUG = !DEBUG
					if DEBUG {
						logrus.SetLevel(logrus.DebugLevel)
						conn.Write([]byte("debug true\n"))
					} else {
						logrus.SetLevel(logrus.InfoLevel)
						logrus.Println("debug false")
						conn.Write([]byte("debug false\n"))
					}
					logrus.Debugln("debug =>", DEBUG)
				} else if IPCMsg != nil {
					rest := IPCMsg(message)
					conn.Write([]byte(rest + "\n"))
				}
			}(conn)
		}
	}()
	return nil
}

func SendMsgToIPC(msg string) (string, error) {
	dial, err := net.Dial("unix", sock)
	if err != nil {
		return "", err
	}
	defer dial.Close()
	dial.Write([]byte(msg + "\n"))
	dial.SetDeadline(time.Now().Add(1 * time.Second))
	reader := bufio.NewReader(dial)
	buf, _, err := reader.ReadLine()
	return string(buf), err
}
func isErrorAddressAlreadyInUse(err error) bool {
	errOpError, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	errSyscallError, ok := errOpError.Err.(*os.SyscallError)
	if !ok {
		return false
	}
	errErrno, ok := errSyscallError.Err.(syscall.Errno)
	if !ok {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}
