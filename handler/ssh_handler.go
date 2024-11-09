package handler

import (
	"fmt"
	"io"
	"trisend/tunnel"
	"trisend/util"

	"github.com/gliderlabs/ssh"
)

func HandleSSH(session ssh.Session) {
	id := util.GetRandomID()
	tunnel.SetStream(id, make(chan tunnel.Stream))
	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s\n", id)

	streamChan, _ := tunnel.GetStream(id)
	stream := <-streamChan

	_, err := io.Copy(stream.Writer, session)
	if err != nil {
		session.Write([]byte("an error occurred while streaming data"))
	}
	close(stream.Done)
}
