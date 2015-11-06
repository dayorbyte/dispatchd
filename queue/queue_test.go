
package queue

import (
  "testing"
)

func TestQueue(t *testing.T) {
  q := NewConcurrentQueue()
  q.PushBack("Hello World!")
  if q.RealLength() != 1 {
    t.Errorf("Length == %d", q.RealLength())
  }
  t.Log("Tail Value: ", q.tail.value)
  value, found := q.PopFront()
  if !found {
    t.Errorf("No value found after PopFront!")
  }
  if value != "Hello World!" {
    t.Error("wrong value found", value)
  }
}