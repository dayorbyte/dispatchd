
package queue

// Lock-free queue data structure based on the paper:
//
// "Simple, Fast, and Practical Non-Blocking and Blocking
//  Concurrent Queue Algorithms" by Maged Michael and Michael Scott
//
// The PDF of the paper is here, hopefully the link will continue to work:
//
// http://www.cs.rochester.edu/u/scott/papers/1996_PODC_queues.pdf
//

import (
  "sync/atomic"
  "unsafe"
)

type NodePointer struct {
}

type ConcurrentQueue struct {
  size int64
  head *Node
  tail *Node
}

type Node struct {
  value interface{}
  next *Node
}

func NewConcurrentQueue() *ConcurrentQueue {
  q := ConcurrentQueue{}
  node := &Node{}
  q.head = node
  q.tail = node
  return &q
}

func (q *ConcurrentQueue) Len() int64 {
  return q.size
}

func (q *ConcurrentQueue) RealLength() (l int64) {
  // NOT THREAD SAFE
  if q.tail == q.head {
    return 0
  }
  for cur := q.tail; cur != nil; cur = cur.next {
    l += 1
  }
  return
}

func (q *ConcurrentQueue) PushBack(value interface{}) {
  node := &Node{value: value}
  var tail *Node =  nil
  for {
    tail = q.tail
    next := tail.next
    if tail == q.tail {
      if next == nil {
        cas := atomic.CompareAndSwapPointer(
          (*unsafe.Pointer)(unsafe.Pointer(&tail.next)),
          unsafe.Pointer(next),
          unsafe.Pointer(node),
        )
        if cas {
          break
        } else {
          atomic.CompareAndSwapPointer(
            (*unsafe.Pointer)(unsafe.Pointer(&q.tail)),
            unsafe.Pointer(tail),
            unsafe.Pointer(next),
          )
        }
      }
    }
  }
  atomic.CompareAndSwapPointer(
    (*unsafe.Pointer)(unsafe.Pointer(&q.tail)),
    unsafe.Pointer(tail),
    unsafe.Pointer(node),
  )
  atomic.AddInt64(&q.size, 1)
}

func (q *ConcurrentQueue) PopFront() (value interface{}, success bool) {
  for {
    head := q.head
    tail := q.tail
    next := head.next
    if head == q.head {
      if head == tail {
        if next == nil {
          return nil, false
        }
        // unsafePointer := unsafe.Pointer(q.tail)
        atomic.CompareAndSwapPointer(
          (*unsafe.Pointer)(unsafe.Pointer(&q.tail)),
          unsafe.Pointer(tail),
          unsafe.Pointer(next),
        )
      } else {
        value = next.value
        // unsafePointer := unsafe.Pointer(q.head)
        cas := atomic.CompareAndSwapPointer(
          (*unsafe.Pointer)(unsafe.Pointer(&q.head)),
          unsafe.Pointer(head),
          unsafe.Pointer(next),
        )
        if cas {
          atomic.AddInt64(&q.size, -1)
          return value, true
        }
      }
    }
  }
}
