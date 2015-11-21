package util

import (
	"regexp"
	"testing"
	"time"
)

func TestRandomId(t *testing.T) {
	var re, err = regexp.Compile("^\\w{32}$")
	if err != nil {
		t.Errorf(err.Error())
	}
	for i := 0; i < 100; i++ {
		var r = RandomId()
		if !re.MatchString(r) {
			t.Errorf("Did not match string: '%s'", r)
		}
	}
}

func TestNextId(t *testing.T) {
	var now = time.Now().UnixNano()
	var next = NextId()
	var nextNext = NextId()
	if now < next {
		t.Errorf("NextId was less than current time")
	}
	if next+1 != nextNext {
		t.Errorf(
			"subsequent NextId calls do not produce incrementing numbers: %d, %d",
			next,
			nextNext,
		)
	}
}
