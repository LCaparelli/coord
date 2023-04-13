package coord

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"
)

type UnixSocketReporter struct {
	conn         net.Conn
	errChan      chan<- error
	isLeaderChan <-chan bool
	isLeaderFunc func() bool
	ticker       *time.Ticker
}

func NewUnixSocketReporter(
	pollInterval time.Duration,
	unixSockPath string,
	errChan chan error,
	isLeaderChan chan bool,
	isLeaderFunc func() bool,
) (UnixSocketReporter, error) {

	conn, err := net.Dial("unix", unixSockPath)
	if err != nil {
		return UnixSocketReporter{}, fmt.Errorf("dialing to socket (%s): %v", unixSockPath, err)
	}

	return UnixSocketReporter{
		conn:         conn,
		errChan:      errChan,
		isLeaderChan: isLeaderChan,
		isLeaderFunc: isLeaderFunc,
		ticker:       time.NewTicker(pollInterval),
	}, nil
}

func (r UnixSocketReporter) Report(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			r.errChan <- r.conn.Close()
			return
		case isLeader := <-r.isLeaderChan:
			_, err := r.conn.Write([]byte(strconv.FormatBool(isLeader)))
			r.errChan <- err
		case <-r.ticker.C:
			_, err := r.conn.Write([]byte(strconv.FormatBool(r.isLeaderFunc())))
			r.errChan <- err
		}
	}
}
