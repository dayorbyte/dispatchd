package amqp

var MaxShortStringLength uint8 = 255
var FrameMethod = 1
var FrameHeader = 2
var FrameBody = 3
var FrameHeartbeat = 8
var FrameMinSize = 4096
var FrameEnd = 206

// Indicates that the method completed successfully. This reply code is
// reserved for future use - the current protocol design does not use positive
// confirmation and reply codes are sent only in case of an error.
var ReplySuccess = 200

// The client attempted to transfer content larger than the server could accept
// at the present time. The client may retry at a later time.
var ContentTooLarge = 311

// When the exchange cannot deliver to a consumer when the immediate flag is
// set. As a result of pending data on the queue or the absence of any
// consumers of the queue.
var NoConsumers = 313

// An operator intervened to close the connection for some reason. The client
// may retry at some later date.
var ConnectionForced = 320

// The client tried to work with an unknown virtual host.
var InvalidPath = 402

// The client attempted to work with a server entity to which it has no
// access due to security settings.
var AccessRefused = 403

// The client attempted to work with a server entity that does not exist.
var NotFound = 404

// The client attempted to work with a server entity to which it has no
// access because another client is working with it.
var ResourceLocked = 405

// The client requested a method that was not allowed because some precondition
// failed.
var PreconditionFailed = 406

// The sender sent a malformed frame that the recipient could not decode.
// This strongly implies a programming error in the sending peer.
var FrameError = 501

// The sender sent a frame that contained illegal values for one or more
// fields. This strongly implies a programming error in the sending peer.
var SyntaxError = 502

// The client sent an invalid sequence of frames, attempting to perform an
// operation that was considered invalid by the server. This usually implies
// a programming error in the client.
var CommandInvalid = 503

// The client attempted to work with a channel that had not been correctly
// opened. This most likely indicates a fault in the client layer.
var ChannelError = 504

// The peer sent a frame that was not expected, usually in the context of
// a content header and body.  This strongly indicates a fault in the peer's
// content processing.
var UnexpectedFrame = 505

// The server could not complete the method because it lacked sufficient
// resources. This may be due to the client creating too many of some type
// of entity.
var ResourceError = 506

// The client tried to work with some entity in a manner that is prohibited
// by the server, due to security settings or by some other criteria.
var NotAllowed = 530

// The client tried to use functionality that is not implemented in the
// server.
var NotImplemented = 540

// The server could not complete the method because of an internal error.
// The server may require intervention by an operator in order to resume
// normal operations.
var InternalError = 541
