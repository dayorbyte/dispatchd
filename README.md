# mq - An AMQP Server Written in Go

## Status

### Wire protcol

* Complete. It can read or write any frame or data type.
* A python script uses the XML spec to generate a lot of what's needed

### Server

* Can perform a simple end to end producer/consumer test on a declared exchange
* Dispatches all received frames into an exchange which then ignores them

### Next Steps

* Full message tracking to allow proper support for ack/nack and confirm mode
* Implemenent the required fanout exchange type
* Add binding support

## Future Goals

* Full compliance with the AMQP specification (plus rabbitmq's extensions)
* Functional and performance testing