package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sourcegraph/conc"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/leaderelection"

	"github.com/lcaparelli/coord/coord"
)

func main() {
	// TODO: add config to parameterize
	namespace := "default"
	electionID := "id"
	leaderPollInterval := time.Second * 10
	unixSockPath := "leader.sock"

	rec := coord.EventRecorderProvider{}

	resLock, err := leaderelection.NewResourceLock(controllerruntime.GetConfigOrDie(), rec, leaderelection.Options{
		LeaderElection:          true,
		LeaderElectionNamespace: namespace,
		LeaderElectionID:        electionID,
	})
	if err != nil {
		panic(fmt.Errorf("creating resource lock: %v", err))
	}

	errChanCoord := make(chan error)
	isLeaderChan := make(chan bool)

	coordinator, err := coord.NewCoordinator(resLock, errChanCoord, isLeaderChan)
	if err != nil {
		panic(fmt.Errorf("creating coordinator: %v", err))
	}

	errChanReport := make(chan error)
	reporter, err := coord.NewUnixSocketReporter(leaderPollInterval, unixSockPath, errChanReport, isLeaderChan, coordinator.IsLeader)
	if err != nil {
		panic(fmt.Errorf("creating reporter: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	var wg conc.WaitGroup
	wg.Go(func() {
		coordinator.Coordinate(ctx)
		close(errChanCoord)
	})
	wg.Go(func() {
		reporter.Report(ctx)
		close(errChanReport)
	})

	wg.Go(func() {
		for err := range errChanCoord {
			if err != nil {
				fmt.Println("coordinator err: " + err.Error())
			}
		}
	})
	wg.Go(func() {
		for err := range errChanReport {
			if err != nil {
				fmt.Println("reporter err: " + err.Error())
			}
		}
	})

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT, syscall.SIGTERM)

	// Block until any signal is received.
	<-c
	cancel()
	wg.Wait()
}
