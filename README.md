# mq - An AMQP Server Written in Go

## Status

* The wire protcol is complete. It can read or write any frame or data type.
  * A python script uses the XML spec to generate a lot of what's needed
  * It seems like it has trouble talking to the rabbitmq perf testing
    tool when a non-empty table is sent. Unclear why since pika is OK with it.
* Other than transactions, all amqp 0.9.1 methods are implemented
* Some rabbitmq extensions are implemented:
  * nack (currently the same consumer could receive the message again)
  * internal exchanges (flag exists, but currently unused)
  * auto-delete exchanges
  * The reinterpretation of basic.qos
* Durability of server settings is supported. Durable messages are not persisted
* No security or auth mechanisms exist (you must use PLAIN auth with
  guest/guest)
* Useful assertions are missing (e.g. declaring a durable auto-delete queue)
* A performance test using rabbitmq's performance testing program showed
  slower throughput for small message sizes and faster throughput for larger
  ones on a direct exchange without acks
* There is a simple admin page that can show basic info about what's
  happening in the server

## Soon

* Durability of messages
* Transactions

## Later

* Github issues: https://github.com/jeffjenkins/mq/issues
* Tests, and probably refactoring to make testing more tractible
* Add flow control so an overactive producer can't swamp the server
* Come up with a real project name
* Performance tuning. The server can do ~20k messages per second