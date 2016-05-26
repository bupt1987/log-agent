package safe

import (
	"container/list"
	"fmt"
	"sync"
)

type Queue struct {
	data *list.List
	lock sync.Mutex
}

func NewQueue() *Queue {
	q := new(Queue)
	q.data = list.New()
	return q
}

func (q *Queue) Push(v interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.data.PushFront(v)
}

func (q *Queue) Pop() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()
	iter := q.data.Back()
	if iter != nil {
		v := iter.Value
		q.data.Remove(iter)
		return v
	}
	return nil
}

func (q *Queue) Dump() {
	q.lock.Lock()
	defer q.lock.Unlock()
	for iter := q.data.Back(); iter != nil; iter = iter.Prev() {
		fmt.Println("item:", iter.Value)
	}
}
