package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type Pool interface {
	Get() (net.Conn, error)
	Put(net.Conn) error
	Close()
	Len() int
}

type NewConn func() (net.Conn, error)

type ConnPool struct {
	newConn NewConn
	mu      sync.RWMutex
	conns   chan net.Conn
}

func NewConnPool(count int, f NewConn) (Pool, error) {
	cp := &ConnPool{
		conns:   make(chan net.Conn, count),
		newConn: f,
	}
	for i := 0; i < count; i++ {
		conn, err := f()
		if err != nil {
			cp.Close()
			return nil, errors.New("Failed to fill the pool, err: " + err.Error())
		}
		cp.conns <- conn
	}
	return cp, nil
}

func (cp *ConnPool) Get() (net.Conn, error) {
	if cp.conns == nil {
		return nil, errors.New("connection is nil")
	}

	// Block here until he gets a conn
	select {
	case conn := <-cp.conns:
		return conn, nil
	}
}

func (cp *ConnPool) Put(conn net.Conn) error {
	if conn == nil {
		return errors.New("connection is nil")
	}

	cp.mu.RLock()
	defer cp.mu.RUnlock()

	select {
	case cp.conns <- conn:
		return nil
	default:
		// pool is full
		return conn.Close()
	}
}

func (cp *ConnPool) Close() {
	cp.mu.Lock()
	conns := cp.conns
	cp.conns = nil
	cp.newConn = nil
	cp.mu.Unlock()
	if conns == nil {
		return
	}
	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (cp *ConnPool) Len() int {
	return len(cp.conns)
}

func main() {
	// Launch server
	go launchTCPserver()

	p, err := NewConnPool(3, func() (net.Conn, error) { return net.Dial(CONN_TYPE, CONN_HOST+":"+CONN_PORT) })
	if err != nil {
		log.Fatal("Failed to new pool, err: ", err)
	}
	fmt.Println("Request count: ", p.Len())

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			conn, err := p.Get()
			if err != nil {
				log.Printf("[Request %d] err: %v\n", i, err)
			} else {
				fmt.Fprintf(conn, "[Request %d] text: %d\n", i, i)
				message, _ := bufio.NewReader(conn).ReadString('\n')
				fmt.Printf("[Request %d] Message from server: %s", i, message)
			}
			if err := p.Put(conn); err != nil {
				log.Printf("[Request %d] err: %v\n", i, err)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	p.Close()
	time.Sleep(1 * time.Second)
}
