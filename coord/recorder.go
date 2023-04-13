package coord

import (
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/recorder"
)

var _ recorder.Provider = EventRecorderProvider{}

type EventRecorderProvider struct{}

// TODO: use a real recorder
func (e EventRecorderProvider) GetEventRecorderFor(name string) record.EventRecorder {
	return &record.FakeRecorder{}
}
