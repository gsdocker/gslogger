package gslogger

import "testing"

func TestGetLogger(t *testing.T) {
	log := Get("Test")

	a := [3]byte{1, 2, 3}

	b := a[:2]

	c := a[:2]

	log.D("b:%d", cap(b))

	b = append(b, 1)

	c = append(c, 2)

	log.D("b:%v", b)

	log.D("c:%v", c)
	log.V("This is VERBOSE message")
	log.D("This is Debug message")
	log.I("This is Text message")
	log.W("This is Warn message")
	log.E("This is Error message")
	log.A("This is Assert message")

	Join()
}
