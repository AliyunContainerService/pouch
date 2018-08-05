package events

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/utils"
)

func TestBasicSubscribe(t *testing.T) {
	type eventsOption struct {
		action    string
		eventType types.EventType
		actor     *types.EventsActor
	}

	ctx := context.Background()
	testEvents := []eventsOption{
		{
			action:    "create",
			eventType: types.EventTypeContainer,
			actor:     &types.EventsActor{ID: "asdf"},
		},
		{
			action:    "create",
			eventType: types.EventTypeContainer,
			actor: &types.EventsActor{
				ID: "qwer",
				Attributes: map[string]string{
					"image": "busybox",
				},
			},
		},
	}

	eventsService := NewEvents()

	t.Log("subscribe")

	// Create two subscribers for same set of events and make sure they
	// traverse the event.
	ctx1, cancel1 := context.WithCancel(ctx)
	eventq1, errq1 := eventsService.Subscribe(ctx1, nil)

	ctx2, cancel2 := context.WithCancel(ctx)
	eventq2, errq2 := eventsService.Subscribe(ctx2, nil)

	t.Log("publish")
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		defer wg.Done()
		defer close(errChan)
		for _, event := range testEvents {
			if err := eventsService.Publish(ctx, event.action, event.eventType, event.actor); err != nil {
				errChan <- err
				return
			}
		}

		t.Log("finished publishing")
	}()

	t.Log("waiting")
	wg.Wait()
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}

	for _, subscriber := range []struct {
		eventq <-chan *types.EventsMessage
		errq   <-chan error
		cancel func()
	}{
		{
			eventq: eventq1,
			errq:   errq1,
			cancel: cancel1,
		},
		{
			eventq: eventq2,
			errq:   errq2,
			cancel: cancel2,
		},
	} {
		var received []types.EventsMessage
	subscribercheck:
		for {
			select {
			case ev := <-subscriber.eventq:
				received = append(received, *ev)
			case err := <-subscriber.errq:
				if err != nil {
					t.Fatal(err)
				}
				break subscribercheck
			}

			if len(received) == 2 {
				// when we do this, we expect the errs channel to be closed and
				// this will return.
				subscriber.cancel()

				for _, ev := range received {
					if ev.Action != "create" || ev.Type != types.EventTypeContainer || !utils.StringInSlice([]string{"asdf", "qwer"}, ev.ID) {
						t.Fatal(fmt.Errorf("got unexpected event message: %#v", ev))
					}
				}
			}
		}
	}
}
