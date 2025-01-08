package server

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"trisend/db"
	"trisend/tunnel"
	"trisend/util"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

const stream_details = "user"

func handlePublicKey(userStore db.UserStore) ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		shaHash := sha256.Sum256(key.Marshal())
		fingerprint := base64.RawStdEncoding.EncodeToString(shaHash[:])

		user, err := userStore.GetBySSHKey(context.Background(), fingerprint)
		if err != nil {
			return true
		}

		streamDetails := &tunnel.StreamDetails{
			Username: user.Username,
			Pfp:      user.Pfp,
		}
		ctx.SetValue(stream_details, streamDetails)

		return true
	}
}

func handleSSH(session ssh.Session) {
	value := session.Context().Value(stream_details)
	if value == nil {
		fmt.Fprintln(session.Stderr(), "You need an Account and register your ssh key")
		session.Exit(1)
	}
	streamDetails := value.(*tunnel.StreamDetails)

	id := util.GetRandomID(10)
	filename := filepath.Base(session.RawCommand())
	noExtName := filename[:len(filename)-len(filepath.Ext(filename))]
	if noExtName == "" {
		noExtName = "compressed_file"
	}

	streamDetails.Filename = noExtName
	tunnel.SetStream(id, make(chan tunnel.Stream), streamDetails)

	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s\n", id)

	streamChan, _ := tunnel.GetStream(id)
	defer close(streamChan)
	stream := <-streamChan

	zipWriter := zip.NewWriter(stream.Writer)
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		fmt.Fprintln(session, "error while transfering data")
		fmt.Printf("ERROR: %s\n", err)
		session.Exit(1)
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

func handleSFTP(userStore db.UserStore) ssh.SubsystemHandler {
	return func(session ssh.Session) {
		shaHash := sha256.Sum256(session.PublicKey().Marshal())
		fingerprint := base64.RawStdEncoding.EncodeToString(shaHash[:])

		user, err := userStore.GetBySSHKey(context.Background(), fingerprint)
		if err != nil {
			fmt.Fprintf(session.Stderr(), "Need to have a registered account\n")
			session.Exit(1)
		}

		handler := &sftpHandler{
			stderr: session.Stderr(),
			streamDetails: &tunnel.StreamDetails{
				Username: user.Username,
				Pfp:      user.Pfp,
			},
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

		err = handler.Close()
		if err != nil {
			fmt.Fprintf(session.Stderr(), "an error occurred while transfering %s\n", err)
			session.Exit(1)
		}
	}
}

type sftpHandler struct {
	sync.Once
	stderr    io.Writer
	zipWriter *zip.Writer

	stream        *tunnel.Stream
	streamDetails *tunnel.StreamDetails
}

func (h *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	h.Do(func() {
		ID := util.GetRandomID(10)
		channel := make(chan tunnel.Stream)

		if h.streamDetails.Filename == "" {
			filename := filepath.Base(r.Filepath)
			noExtName := filename[:len(filename)-len(filepath.Ext(filename))]

			h.streamDetails.Filename = noExtName
		}

		tunnel.SetStream(ID, channel, h.streamDetails)
		fmt.Fprintf(h.stderr, "LINK: http://localhost:3000/download/%s\n", ID)

		stream := <-channel
		h.stream = &stream
		h.zipWriter = zip.NewWriter(h.stream.Writer)
	})

	fileWriter, err := h.zipWriter.Create(filepath.Base(r.Filepath))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, fmt.Errorf("error while transfering data\n")
	}

	var writeAt writerAt
	writeAt.writer = fileWriter

	return &writeAt, nil
}

func (h *sftpHandler) Close() error {
	close(h.stream.Done)
	if err := h.zipWriter.Close(); err != nil {
		return err
	}

	return nil
}

func (h *sftpHandler) Filecmd(r *sftp.Request) error {
	// it executes only if it is a directoy before transfer
	if r.Method == "Mkdir" {
		h.streamDetails.Filename = filepath.Base(r.Filepath)
		return nil
	}
	// it executes after transfer
	if r.Method == "Setstat" {
		return nil
	}
	return sftp.ErrSshFxOpUnsupported
}

type writerAt struct {
	writer io.Writer
}

func (h *writerAt) WriteAt(p []byte, off int64) (n int, err error) {
	return h.writer.Write(p)
}
