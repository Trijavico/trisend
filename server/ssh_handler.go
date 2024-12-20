package server

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"trisend/tunnel"
	"trisend/util"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

func handleSSH(session ssh.Session) {
	id := util.GetRandomID(10)
	filename := filepath.Base(session.RawCommand())
	noExtName := filename[:len(filename)-len(filepath.Ext(filename))]
	if noExtName == "" {
		noExtName = "compressed_file"
	}
	tunnel.SetStream(id, make(chan tunnel.Stream))
	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s?zip=%s\n", id, noExtName)

	streamChan, _ := tunnel.GetStream(id)
	defer close(streamChan)
	stream := <-streamChan

	zipWriter := zip.NewWriter(stream.Writer)
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		fmt.Fprintln(session, "error while transfering data")
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	_, err = io.Copy(fileWriter, session)
	if err != nil {
		session.Write([]byte("an error occurred while streaming data"))
	}

	if err := zipWriter.Close(); err != nil {
		fmt.Fprintf(session.Stderr(), "error closing zip: %s\n", err)
	}
	close(stream.Done)
}

func handleSFTP(session ssh.Session) {
	ID := util.GetRandomID(10)
	streamChan := make(chan tunnel.Stream)
	tunnel.SetStream(ID, streamChan)
	defer close(streamChan)

	handler := &sftpHandler{
		id:         ID,
		session:    session,
		streamChan: streamChan,
	}

	sftpHandlers := sftp.Handlers{
		FileGet:  sftp.InMemHandler().FileGet,
		FilePut:  handler,
		FileCmd:  handler,
		FileList: sftp.InMemHandler().FileList,
	}

	srv := sftp.NewRequestServer(session, sftpHandlers)
	if err := srv.Serve(); err != nil && err != io.EOF {
		fmt.Fprintf(session.Stderr(), "an error occurred while transfering %s\n", err)
	}

	if handler.writeStream == nil {
		return
	}

	if err := handler.zipWriter.Close(); err != nil {
		fmt.Fprintf(session.Stderr(), "error closing zip: %s\n", err)
	}

	close(handler.writeStream.Done)
}

type sftpHandler struct {
	id          string
	queryParam  string
	session     ssh.Session
	streamChan  chan tunnel.Stream
	writeStream *tunnel.Stream
	zipWriter   *zip.Writer
}

type writerAt struct {
	writer io.Writer
}

func (h *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	return h.writer.Write(p)
}

func (h *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	if h.writeStream == nil {
		if h.queryParam == "" {
			filename := filepath.Base(r.Filepath)
			noExtName := filename[:len(filename)-len(filepath.Ext(filename))]
			h.queryParam = fmt.Sprintf("?zip=%s", noExtName)
		}
		fmt.Fprintf(h.session.Stderr(), "LINK: http://localhost:3000/download/%s%s\n", h.id, h.queryParam)
		writeStream := <-h.streamChan
		h.writeStream = &writeStream
	}
	if h.zipWriter == nil {
		h.zipWriter = zip.NewWriter(h.writeStream.Writer)
	}

	fileWriter, err := h.zipWriter.Create(filepath.Base(r.Filepath))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, fmt.Errorf("error while transfering data\n")
	}
	writeAt := writerAt{
		writer: fileWriter,
	}

	return &writeAt, nil
}

func (h *sftpHandler) Filecmd(r *sftp.Request) error {
	// it executes only if it is a directoy before transfer
	if r.Method == "Mkdir" {
		h.queryParam = fmt.Sprintf("?zip=%s", filepath.Base(r.Filepath))
		return nil
	}
	// it executes after transfer
	if r.Method == "Setstat" {
		fmt.Fprintf(os.Stdin, "DEBUGGIN!!!\n")
		return nil
	}
	return sftp.ErrSshFxOpUnsupported
}
