package main

// Soft error (close channel)

type AMQPError struct {
	Code   uint16
	Class  uint16
	Method uint16
	Msg    string
	Soft   bool
}

func NewSoftError(code uint16, msg string, class uint16, method uint16) *AMQPError {
	return &AMQPError{
		Code:   code,
		Class:  class,
		Method: method,
		Msg:    msg,
		Soft:   true,
	}
}

// Hard error (close connection)

func NewHardError(code uint16, msg string, class uint16, method uint16) *AMQPError {
	return &AMQPError{
		Code:   code,
		Class:  class,
		Method: method,
		Msg:    msg,
		Soft:   false,
	}
}
