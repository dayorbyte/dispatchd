# mq - An AMQP Server Written in Go

## Status

### Wire protcol

* Complete. It can read or write any frame or data type.
* A python script uses the XML spec to generate a lot of what's needed

### Server

* Basic functionality to publish and consume messages, create
  queues/exchanges exists.
* A few functions aren't implemented:
  * flow control
  * synchronous get from a queue
  * return
  * recover
  * cancel consumer
  * connection blocked
  * exchange to exchange binding (protocol extension)
* None of the life cycle parts of the spec are implemented:
  * durability of anything
  * immediate and mandatory
* No security or auth mechanisms exist (you must use PLAIN auth with
  guest/guest)
* Lots of assertions are missing (key format, message sizes)
* The publisher confirmation extension is largely unimplemented
* A performance test using rabbitmq's performance testing program showed
  slower throughput for small message sizes and faster throughput for larger
  ones on a direct exchange without acks
* Transactions are not supported and there is no plan to support them
* There is a simple admin page that can show basic info about what's
  happening in the server

### Next Steps

* Durability of server configuration (but not messages, yet)
* Implement the remaining methods from the spec listed above
* Add the rest of the assertions from the spec
* Add flow control so an overactive producer can't swamp the server
* Come up with a real project name