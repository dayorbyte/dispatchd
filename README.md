# dispatchd - A Message Broker and Queue Server

## Status

`dispatchd` is in alpha.

It generally works but is not hardened enough for production use. More on what it can and can't do below

## What's done?

* The wire protcol (amqp 0-9-1) is complete. It can read or write any frame or data type.
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
* Transactions are supported
* Durability of messages is supported, although currently buggy

## Next

* More tests. I'm aiming for full test coverage. I'm about ~1/3 of the way
  there and all work right now is writing tests
* Faster persistent messages. Right now it's like ~200qps with peristence
  on. I also think I can not persist messages which I can send off right
  away if the response is fast enough
* Faster consumption. I can't seem to get more than 8000qps from a single
  consumer, but I can do like 30000 with producers

## Later

* Github issues: https://github.com/jeffjenkins/mq/issues
* Tests, and probably refactoring to make testing more tractible
* Add flow control so an overactive producer can't swamp the server
* Come up with a real project name
* Performance tuning. The server can do ~20k messages per second with persistence turned off.