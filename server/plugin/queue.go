package plugin

import (
	"container/heap"
	"context"
	"github.com/mattermost/mattermost-plugin-pomodoro/server/model"
	"github.com/pkg/errors"
	"time"
)

var tickTime = 5 * time.Second

func (p *Plugin) NewWorkQueue(workers int) *SessionQueue {
	sessionChan := make(chan *model.Session)
	ctx, cancel := context.WithCancel(context.Background())
	waitChan := make(chan *delayedSession)

	for i := 0; i < workers; i++ {
		go p.worker(ctx, sessionChan, waitChan)
	}

	queue := &SessionQueue{
		ctx:            ctx,
		ctxCancel:      cancel,
		heartbeat:      time.NewTicker(tickTime),
		workersCount:   workers,
		sessionChannel: sessionChan,
		waitChan:       waitChan,
	}

	go queue.runWaitLoop()

	return queue
}

type SessionQueue struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	heartbeat *time.Ticker

	workersCount   int
	sessionChannel chan *model.Session // TODO: consider enqueuing user ids instead, as one user can have only one active session

	waitChan chan *delayedSession
}

func (q *SessionQueue) Cancel() {
	q.ctxCancel()
}

func (q *SessionQueue) Add(session *model.Session) {
	q.sessionChannel <- session
}

func (p *Plugin) worker(ctx context.Context, sessionChan <-chan *model.Session, waitChan chan<- *delayedSession) {
	for {
		select {
		case <-ctx.Done():
			return

		case session := <-sessionChan:
			done, err := p.processSession(session)
			if err != nil {
				p.API.LogError("Failed to process session, requeuing", "error", err)
				waitChan <- &delayedSession{session: session, readyAt: time.Now().Unix() + 1} // Add one second delay if error occurred
			} else if !done {
				//p.API.LogDebug("Session not finished, requeuing")
				waitChan <- &delayedSession{session: session, readyAt: session.StartTime + session.Length}
			}

			if ctx.Err() != nil {
				p.API.LogDebug("Worker exiting")
				return
			}
		}
	}
}

func (p *Plugin) processSession(session *model.Session) (bool, error) {
	endTime := session.StartTime + session.Length

	if time.Now().Unix() < endTime {
		return false, nil
	}

	err := p.finalizeSession(session.UserID)
	if err != nil {
		return false, errors.Wrap(err, "Failed to finalize session")
	}

	return true, nil
}

func (q *SessionQueue) runWaitLoop() {
	queue := &delayedSessionPriorityQueue{}
	heap.Init(queue)

	var nextItemChannel *time.Timer

	for {
		now := time.Now().Unix()

		// Add ready entries
		for queue.Len() > 0 {
			entry := queue.Peek().(*delayedSession)
			// First item is not ready, no need to check further
			if entry.readyAt > now {
				break
			}

			entry = heap.Pop(queue).(*delayedSession)
			q.Add(entry.session)
		}

		// If not next item, this will never return
		nextReadyAt := make(<-chan time.Time)
		if queue.Len() > 0 {
			if nextItemChannel != nil {
				nextItemChannel.Stop()
			}

			entry := queue.Peek().(*delayedSession)
			nextItemChannel = time.NewTimer(time.Duration(now-entry.readyAt) * time.Second)
			nextReadyAt = nextItemChannel.C
		}

		select {
		case <-nextReadyAt:
			// next item is ready, continue loop to add it to work queue
		case <-q.heartbeat.C:
			// continue loop - check if next item is ready
		case entry := <-q.waitChan:
			endTime := entry.session.StartTime + entry.session.Length
			if endTime < time.Now().Unix() {
				q.Add(entry.session)
				continue
			}

			queue.Push(entry)
		}
	}
}
