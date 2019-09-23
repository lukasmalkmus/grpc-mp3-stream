package main

import (
	"context"
	"io"
	"log"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"google.golang.org/grpc"

	"github.com/lukasmalkmus/stream/bytestream"
)

const (
	remoteAddr   = "localhost:10000"
	resourceName = "song.mp3"
)

func main() {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		log.Printf("Dial(%q): %v", remoteAddr, err)
		return
	}

	client := bytestream.NewClient(conn)
	reader, err := client.NewReader(context.Background(), resourceName)
	if err != nil {
		log.Printf("NewReader(%q): %v", resourceName, err)
		return
	}

	dec, err := mp3.NewDecoder(reader)
	if err != nil {
		log.Printf("NewDecoder(): %v", err)
		return
	}

	p, err := oto.NewPlayer(dec.SampleRate(), 2, 2, bytestream.MaxBufSize)
	if err != nil {
		log.Printf("NewPlayer(): %v", err)
		return
	}
	defer p.Close()

	if _, err := io.Copy(p, dec); err != nil {
		log.Printf("Copy(): %v", err)
	}
}
