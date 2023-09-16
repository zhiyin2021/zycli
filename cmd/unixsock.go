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
	IPCMsg func(msg string) string
	// sock      = tools.CurrentName() + ".ipc"
	ulistener net.Listener
)

func stopUnixSock() {
	if ulistener != nil {
		os.Remove(defOpt.ipcPath)
	}
}
func IsRuning() bool {
	p := tools.FixPath(defOpt.ipcPath)
	dial, err := net.Dial("unix", p)
	if err == nil {
		dial.Close()
	} else {
		os.Remove(p)
	}
	return err == nil
}
func startUnixSock() error {
	// addr, _ := net.ResolveUnixAddr("unix", sock)
	var err error
	ulistener, err = net.Listen("unix", defOpt.ipcPath)
	if err != nil {
		if isErrorAddressAlreadyInUse(err) {
			logrus.Errorf("please check application already running.")
		} else {
			logrus.Errorln("usock", err)
		}
		return err
	}
	go func() {
		defer OnPanic(nil)
		defer ulistener.Close()
		for {
			conn, err := ulistener.Accept()
			if err != nil {
				continue
			}
			go func(con net.Conn) {
				defer con.Close()
				reader := bufio.NewReader(conn)
				msg, err := reader.ReadSlice(0)
				if err != nil {
					return
				}
				logrus.Println("uread:", msg)
				message := string(msg[:len(msg)-1])
				if message == "dbg" {
					DEBUG = !DEBUG
					if DEBUG {
						logrus.SetLevel(logrus.DebugLevel)
						conn.Write([]byte("debug true\x00"))
					} else {
						logrus.SetLevel(logrus.InfoLevel)
						logrus.Println("debug false")
						conn.Write([]byte("debug false\x00"))
					}
					logrus.Debugln("debug =>", DEBUG)
				} else if IPCMsg != nil {
					rest := IPCMsg(message)
					conn.Write([]byte(rest + "\x00"))
				}
			}(conn)
		}
	}()
	return nil
}

func SendMsgToIPC(msg string) (string, error) {
	defer OnPanic(nil)
	dial, err := net.Dial("unix", defOpt.ipcPath)
	if err != nil {
		return "", err
	}
	defer dial.Close()
	dial.Write([]byte(msg + "\x00"))
	dial.SetDeadline(time.Now().Add(5 * time.Second))

	reader := bufio.NewReaderSize(dial, 1024*1024)

	buf, err := reader.ReadSlice(0)
	if len(buf) > 0 {
		buf = buf[:len(buf)-1]
	}
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
