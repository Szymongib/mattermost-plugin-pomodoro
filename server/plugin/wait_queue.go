package plugin

import (
	"github.com/mattermost/mattermost-plugin-pomodoro/server/model"
)

type delayedSession struct {
	session *model.Session
	readyAt int64
	index   int
}

// delayedSessionPriorityQueue implements heap.Interface. It ensures that elements with smallest
// readyAt will be at the top.
type delayedSessionPriorityQueue []*delayedSession

func (pq delayedSessionPriorityQueue) Len() int {
	return len(pq)
}
func (pq delayedSessionPriorityQueue) Less(i, j int) bool {
	return pq[i].readyAt < pq[j].readyAt
}
func (pq delayedSessionPriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds an item to the queue. Push should not be called directly; instead,
// use `heap.Push`.
func (pq *delayedSessionPriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*delayedSession)

	// Avoid duplicated entries
	for _, s := range *pq {
		if s.session.UserID == item.session.UserID {
			return
		}
	}

	item.index = n
	*pq = append(*pq, item)
}

// Pop removes an item from the queue. Pop should not be called directly;
// instead, use `heap.Pop`.
func (pq *delayedSessionPriorityQueue) Pop() interface{} {
	n := len(*pq)
	item := (*pq)[n-1]
	item.index = -1
	*pq = (*pq)[0:(n - 1)]
	return item
}

// Peek returns the item at the beginning of the queue, without removing the
// item or otherwise mutating the queue. It is safe to call directly.
func (pq delayedSessionPriorityQueue) Peek() interface{} {
	return pq[0]
}
