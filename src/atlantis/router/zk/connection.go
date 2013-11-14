package zk

import (
	"errors"
	"launchpad.net/gozk"
	"log"
	"sync"
	"time"
)

type ZkConn struct {
	sync.Mutex
	ResetCh chan bool
	servers string
	Conn    *zookeeper.Conn
	eventCh <-chan zookeeper.Event
	killCh  chan bool
}

func ManagedZkConn(servers string) *ZkConn {
	zk := &ZkConn{
		ResetCh: make(chan bool),
		servers: servers,
		killCh:  make(chan bool),
	}

	go zk.dialExclusive()

	return zk
}

func (z *ZkConn) Shutdown() {
	z.killCh <- true
	z.Conn.Close()
}

func (z *ZkConn) dialExclusive() {
	z.Lock()

	for err := z.dial(); err != nil; {
		log.Printf("z.dial(): %s", err)
	}

	z.Unlock()

	z.ResetCh <- true
}

func (z *ZkConn) dial() error {
	var err error
	z.Conn, z.eventCh, err = zookeeper.Dial(z.servers, 30*time.Second)
	if err != nil {
		return err
	}

	err = z.waitOnConnect()
	if err != nil {
		return err
	}

	go z.monitorEventCh()

	return nil
}

func (z *ZkConn) waitOnConnect() error {
	for {
		ev := <-z.eventCh
		log.Printf("waitOnConnect: %d %s", ev.State, ev)

		switch ev.State {
		case zookeeper.STATE_CONNECTED:
			return nil
		case zookeeper.STATE_CONNECTING:
			continue
		default:
			return errors.New(ev.String())
		}
	}
}

func (z *ZkConn) monitorEventCh() {
	for {
		select {
		case ev := <-z.eventCh:
			log.Printf("monitorEventCh: %d, %s", ev.State, ev)
			if ev.State == zookeeper.STATE_EXPIRED_SESSION ||
				ev.State == zookeeper.STATE_CONNECTING {
				z.dialExclusive()
				return
			}

		case <-z.killCh:
			return
		}
	}
}
