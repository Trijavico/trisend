package handler

import (
	"fmt"
	"io"
	"trisend/tunnel"
	"trisend/util"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

func HandleSSH(session ssh.Session) {
	id := util.GetRandomID()
	tunnel.SetStream(id, make(chan tunnel.Stream))
	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s\n", id)

	streamChan, _ := tunnel.GetStream(id)
	defer close(streamChan)
	stream := <-streamChan

	_, err := io.Copy(stream.Writer, session)
	if err != nil {
		session.Write([]byte("an error occurred while streaming data"))
	}
	close(stream.Done)
}

func HandleSFTP(session ssh.Session) {
	var handler sftpHandler

	sftpHandlers := sftp.Handlers{
		FileGet:  sftp.InMemHandler().FileGet,
		FilePut:  handler,
		FileCmd:  handler,
		FileList: sftp.InMemHandler().FileList,
	}

	srv := sftp.NewRequestServer(session, sftpHandlers)

	if err := srv.Serve(); err != nil {
		fmt.Fprintf(session, "an error occurred while transfering")
	}
}

type sftpHandler struct{}

type writerAt struct {
	writer io.Writer
}

func (t *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	return t.writer.Write(p)
}

func (t sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	ID := util.GetRandomID()
	fmt.Printf("LINK: http://localhost:3000/download/%s\n", ID)

	tunnel.SetStream(ID, make(chan tunnel.Stream))
	streamChan, _ := tunnel.GetStream(ID)
	stream := <-streamChan

	writeAt := writerAt{
		writer: stream.Writer,
	}

	go func() {
		<-r.Context().Done()
		close(streamChan)
		close(stream.Done)
	}()

	return &writeAt, nil
}

func (t sftpHandler) Filecmd(r *sftp.Request) error {
	if r.Method == "Setstat" {
		return nil
	}
	return sftp.ErrSshFxOpUnsupported
}
