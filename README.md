# mq - An AMQP Server Written in Go

## Status

### Wire protcol

* Mostly complete.
* A python script uses the XML spec to generate a lot of what's needed
* Helpers are available to read/write most types.
* Notably missing: writing Table fields and dealing with packing multiple bits

### Server

* Can go through the whole client handshake using the plain auth mechanism
* Dispatches all received frames into a function which throws them away


## Future Goals

* Full compliance with the AMQP specification (plus rabbitmq's extensions)
* Functional and performance testing