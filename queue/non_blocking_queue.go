
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

type Queue struct {
  head *Node
  tail *Node
}

type Node struct {
  value interface{}
  next *Node
}

func (q *Queue) PushBack(value interface{}) {
  node := &Node{value: value}
  node.next = nil
  var tail *Node =  nil
  for {
    tail = q.tail
    next := tail.next
    if tail == q.tail {
      if next == nil {
        unsafePointer := unsafe.Pointer(tail.next)
        cas := atomic.CompareAndSwapPointer(
          &unsafePointer,
          unsafe.Pointer(next),
          unsafe.Pointer(node),
        )
        if cas {
          break
        } else {
          unsafePointer := unsafe.Pointer(q.tail)
          atomic.CompareAndSwapPointer(
            &unsafePointer,
            unsafe.Pointer(tail),
            unsafe.Pointer(next),
          )
        }
      }
    }
  }
  unsafePointer := unsafe.Pointer(q.tail)
  atomic.CompareAndSwapPointer(
    &unsafePointer,
    unsafe.Pointer(tail),
    unsafe.Pointer(node),
  )
}

func (q *Queue) PopFront() (value interface{}, success bool) {
  for {
    head := q.head
    tail := q.tail
    next := head.next
    if head == q.head {
      if head == tail {
        if next == nil {
          return nil, false
        }
        unsafePointer := unsafe.Pointer(q.tail)
        atomic.CompareAndSwapPointer(
          &unsafePointer,
          unsafe.Pointer(tail),
          unsafe.Pointer(next),
        )
      } else {
        value = next.value
        unsafePointer := unsafe.Pointer(q.head)
        cas := atomic.CompareAndSwapPointer(
          &unsafePointer,
          unsafe.Pointer(head),
          unsafe.Pointer(next),
        )
        if cas {
          break
        }
      }
    }
  }
  return
}