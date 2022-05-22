package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"github.com/yglcode/echo_any_port/tube_sock"
)

type EchoServer struct {
	appName  string
	ports    []int
	locPort  int
	listener *net.TCPListener
	sync.Mutex
	conns    map[*net.TCPConn]struct{}
	wg       sync.WaitGroup
}

func main() {
	portStrs := os.Args[1:]
	if len(portStrs) < 1 {
		log.Fatalln("usage: echo port,...")
	}
	ports := []int{}
	for _, s := range portStrs {
		if pn, err := strconv.Atoi(s); err != nil {
			log.Fatalf("invalid port: %s", s)
		} else {
			ports = append(ports, pn)
		}
	}

	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, os.Interrupt)

	server, err := newEchoServer(ports)
	if err != nil {
		log.Fatalf("failed to start server, error: %v", err)
	}
	server.Start()

	//wait for shutdown signal
	select {
	case <-doneCh:
		log.Println("shutdown signal recieved")
	}

	server.Close()
	log.Println("echo server exit...")
}

func newEchoServer(ports []int) (srv *EchoServer, err error) {
	srv = &EchoServer{
		appName: "tubular-echo-server",
		ports:   ports,
		conns: make(map[*net.TCPConn]struct{}),
	}
	//let use random local tcp port
	l, err := net.ListenTCP("tcp4", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Println(err)
		return
	}
	srv.listener = l
	srv.locPort = l.Addr().(*net.TCPAddr).Port
	log.Printf("listener at local addr: %v, port: %d", l.Addr().(*net.TCPAddr), srv.locPort)
	return
}

func (srv *EchoServer) Start() {
	var waitStart sync.WaitGroup
	log.Printf("listener start: %v\n", srv.listener)
	waitStart.Add(1)
	go func(l *net.TCPListener) {
		waitStart.Done()
		for {
			conn, err := srv.listener.AcceptTCP()
			if err != nil {
				//server shutdown
				log.Println(err)
				break
			}
			srv.Lock()
			srv.conns[conn]=struct{}{}
			srv.Unlock()
			log.Printf("new conn: %v", conn)
			//run echo loop
			go srv.echoLoop(conn)
		}
		log.Printf("listener exit: %v...\n", l)
	}(srv.listener)
	//have to wait for listener to start accepting, for BPF to bind to
	waitStart.Wait()
	time.Sleep(500 * time.Millisecond)
	//enabl bpf sock map
	tube_sock.Start(srv.appName, srv.locPort, srv.ports)
}

func (srv *EchoServer) Close() {
	tube_sock.Stop(srv.appName, srv.ports)
	srv.listener.Close()
	// ask echo loops to exit
	srv.Lock()
	for c,_ := range srv.conns {
		c.Close()
	}
	srv.Unlock()
	// wait for echo loops exit
	srv.wg.Wait()
}

func (srv *EchoServer) echoLoop(conn *net.TCPConn) {
	log.Printf("echoLoop start for conn: %v", conn)
	srv.wg.Add(1)
	if _, err := io.Copy(conn, conn); err != nil {
		log.Printf("echoLoop exit error: %v", err)
	}
	log.Printf("echoLoop exit...")
	srv.Lock()
	delete(srv.conns, conn)
	srv.Unlock()
	srv.wg.Done()
}
