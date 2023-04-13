package coord

import (
	"context"
	"time"

	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Coordinator struct {
	errChan      chan<- error
	elector      *leaderelection.LeaderElector
	isLeaderChan chan<- bool
}

func NewCoordinator(resourceLock resourcelock.Interface, errChan chan<- error, isLeaderChan chan<- bool) (Coordinator, error) {
	c := Coordinator{
		errChan:      errChan,
		isLeaderChan: isLeaderChan,
	}

	// TODO: add a config to parameterize all of this
	elector, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: time.Second * 15,
		RenewDeadline: time.Second * 10,
		RetryPeriod:   time.Second * 2,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ context.Context) {
				c.isLeaderChan <- true
			},
			OnStoppedLeading: func() {
				c.isLeaderChan <- false
			},
		},
		WatchDog:        nil, // TODO: check if this is useful when implementing health check, but it's likely more useful to the app itself
		ReleaseOnCancel: false,
		Name:            "coordinator",
	})
	if err != nil {
		return Coordinator{}, err
	}

	c.elector = elector

	return c, nil
}

func (c Coordinator) Coordinate(ctx context.Context) {
	c.elector.Run(ctx)
}

func (c Coordinator) IsLeader() bool {
	return c.elector.IsLeader()
}
