# mq - An AMQP Server Written in Go

## Status

### Wire protcol

* Mostly complete.
* A python script uses the XML spec to generate a lot of what's needed


### Server

* Can go through the whole client handshake using the plain auth mechanism
* Dispatches all received frames into an exchange which then ignores them

### Next Steps

* Finishing up implementing the connection/channel error handling mechanisms
* Implemenent the required direct and fanout exchanges
* Implement enough consumption functions to do an end-to-end test

## Future Goals

* Full compliance with the AMQP specification (plus rabbitmq's extensions)
* Functional and performance testing