# lukasmalkmus/stream

> A quick mp3 streaming example.

---

I put this together to really quickly demonstrate streaming a mp3 file using
gRPC. This example doesn't just chunk the file and send it to the client before
decoding and playing it, it really "streams" it by utilizing Go's powerful
io.Reader interface. The `bytestream` package provides a io.Reader
implementation on top of gRPC.

Package `bytestream` is essentially https://godoc.org/google.golang.org/api/transport/bytestream
but trimmed down for simplicity. It omits all the io.Writer bits.

`bytestream.proto` is essentially https://github.com/googleapis/googleapis/blob/master/google/bytestream/bytestream.proto but trimmed down as well.
