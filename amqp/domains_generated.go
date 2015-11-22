package amqp

var ReadMethodId = ReadShort
var WriteMethodId = WriteShort
var ReadReplyText = ReadShortstr
var WriteReplyText = WriteShortstr
var ReadPeerProperties = ReadTable
var WritePeerProperties = WriteTable
var ReadClassId = ReadShort
var WriteClassId = WriteShort
var ReadMessageCount = ReadLong
var WriteMessageCount = WriteLong
var ReadReplyCode = ReadShort
var WriteReplyCode = WriteShort
var ReadQueueName = ReadShortstr
var WriteQueueName = WriteShortstr
var ReadConsumerTag = ReadShortstr
var WriteConsumerTag = WriteShortstr
var ReadPath = ReadShortstr
var WritePath = WriteShortstr
var ReadExchangeName = ReadShortstr
var WriteExchangeName = WriteShortstr
var ReadDeliveryTag = ReadLonglong
var WriteDeliveryTag = WriteLonglong
