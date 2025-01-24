package server

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
	"trisend/db"
	"trisend/tunnel"
	"trisend/util"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

const (
	stream_details = "user"
	limit          = 5295309 // 5.05MB
	timeout        = time.Minute * 10
)

var (
	maxLimitError   = fmt.Errorf("Limit REACHED: 5.04 MB")
	defaultError    = fmt.Errorf("An error has occurred, try it later.")
	authError       = fmt.Errorf("No Account found with SSH key. Create a new account.")
	expirationError = fmt.Errorf("15 minutes has been expired")
)

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
	errChanClosed := true

	value := session.Context().Value(stream_details)
	if value == nil {
		fmt.Fprintln(session.Stderr(), authError)
		session.Exit(1)
		return
	}
	streamDetails := value.(*tunnel.StreamDetails)

	id := util.GetRandomID(10)
	filename := filepath.Base(session.RawCommand())
	noExtName := filename[:len(filename)-len(filepath.Ext(filename))]
	if noExtName == "" || filename == "" {
		fmt.Fprintln(session.Stderr(), "ssh trisend <filename> < <filepath>")
		session.Exit(1)
		return
	}

	temp, err := os.CreateTemp("", "trisend-*.temp")
	if err != nil {
		slog.Error(err.Error())
		fmt.Fprintln(session, defaultError)
		session.Exit(1)
		return
	}
	defer temp.Close()

	streamDetails.Filename = noExtName
	streamDetails.Expires = time.Now().Add(timeout)
	tunnel.SetStream(id, make(chan tunnel.Stream), streamDetails)

	fmt.Fprintf(session, "LINK: http://localhost:3000/download/%s\n", id)

	channel, _ := tunnel.GetStream(id)
	defer close(channel)

	var stream tunnel.Stream
	select {
	case stream = <-channel:
	case <-time.After(timeout):
		fmt.Fprintln(session.Stderr(), expirationError)
		tunnel.DeleteStream(id)
		session.Exit(1)
		return
	}

	defer func() {
		close(stream.Done)
		if !errChanClosed {
			close(stream.Error)
		}
	}()

	limitReader := io.LimitReader(session, limit)

	amount, err := io.Copy(temp, limitReader)
	if err != nil {
		close(stream.Error)
		fmt.Fprintln(session.Stderr(), maxLimitError)
		session.Exit(1)
		return
	}
	if amount == limit {
		close(stream.Error)
		fmt.Fprintln(session, maxLimitError)
		session.Exit(1)
		return
	}

	_, err = temp.Seek(0, 0)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	stream.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", noExtName))
	zipWriter := zip.NewWriter(stream.Writer)
	fileWriter, err := zipWriter.Create(filename)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	errChanClosed = false
	io.Copy(fileWriter, temp)
	zipWriter.Close()
}

func handleSFTP(userStore db.UserStore) ssh.SubsystemHandler {
	return func(session ssh.Session) {
		errChanClosed := true

		shaHash := sha256.Sum256(session.PublicKey().Marshal())
		fingerprint := base64.RawStdEncoding.EncodeToString(shaHash[:])

		user, err := userStore.GetBySSHKey(context.Background(), fingerprint)
		if err != nil {
			fmt.Fprintln(session.Stderr(), authError)
			session.Exit(1)
			return
		}

		temp, err := os.CreateTemp("", "trisend-*.temp")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		defer temp.Close()

		streamDetails := new(tunnel.StreamDetails)
		streamDetails.Username = user.Username
		streamDetails.Pfp = user.Pfp

		expiration := time.Now().Add(timeout)
		streamDetails.Expires = expiration

		handler := newSFTPHandler(
			session.Stderr(),
			temp,
			streamDetails,
		)
		defer func() {
			if time.Now().After(expiration) {
				return
			}
			close(handler.stream.Done)
			if !errChanClosed {
				close(handler.stream.Error)
			}
		}()

		srv := sftp.NewRequestServer(session, handler.Build())
		handler.server = srv

		if err := srv.Serve(); err != nil && err != io.EOF {
			if time.Now().After(expiration) {
				fmt.Fprintln(session.Stderr(), expirationError)
				session.Exit(1)
				return
			}

			close(handler.stream.Error)
			if errors.Is(err, maxLimitError) {
				fmt.Fprintln(session.Stderr(), maxLimitError)
				session.Exit(1)
				return
			}

			slog.Error(err.Error())
			fmt.Fprintln(session, defaultError)
			session.Exit(1)
			return
		}

		err = handler.zipWriter.Close()
		if err != nil {
			close(handler.stream.Error)
			slog.Error(err.Error())
			fmt.Fprintln(session.Stderr(), defaultError)
			session.Exit(1)
			return
		}

		_, err = handler.tempFile.Seek(0, 0)
		if err != nil {
			close(handler.stream.Error)
			slog.Error(err.Error())
			fmt.Fprintln(session.Stderr(), defaultError)
			session.Exit(1)
			return
		}

		errChanClosed = false
		handler.stream.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", streamDetails.Filename))

		io.Copy(handler.stream.Writer, handler.tempFile)
	}
}

type sftpHandler struct {
	sync.Once
	stderr     io.Writer
	zipWriter  *zip.Writer
	totalSize  *int
	tempFile   *os.File
	server     *sftp.RequestServer
	fileWriter io.Writer

	stream        *tunnel.Stream
	streamDetails *tunnel.StreamDetails
}

func newSFTPHandler(stderr io.ReadWriter, temp *os.File, streamDetails *tunnel.StreamDetails) *sftpHandler {
	return &sftpHandler{
		stderr:        stderr,
		tempFile:      temp,
		totalSize:     new(int),
		streamDetails: streamDetails,
	}
}

func (h *sftpHandler) Build() sftp.Handlers {
	return sftp.Handlers{
		FileGet:  sftp.InMemHandler().FileGet,
		FilePut:  h,
		FileCmd:  h,
		FileList: sftp.InMemHandler().FileList,
	}
}

func (h *sftpHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	h.Do(func() {
		ID := util.GetRandomID(10)
		channel := make(chan tunnel.Stream)
		defer close(channel)

		if h.streamDetails.Filename == "" {
			filename := filepath.Base(r.Filepath)
			noExtName := filename[:len(filename)-len(filepath.Ext(filename))]

			h.streamDetails.Filename = noExtName
		}

		h.streamDetails.Expires = time.Now().Add(timeout)
		tunnel.SetStream(ID, channel, h.streamDetails)
		fmt.Fprintf(h.stderr, "LINK: http://localhost:3000/download/%s\n", ID)

		select {
		case stream := <-channel:
			h.stream = &stream
		case <-time.After(timeout):
			fmt.Fprintln(h.stderr, expirationError)
			tunnel.DeleteStream(ID)
			h.server.Close()
		}

		h.zipWriter = zip.NewWriter(h.tempFile)
	})

	fileWriter, err := h.zipWriter.Create(filepath.Base(r.Filepath))
	if err != nil {
		slog.Error(err.Error())
		return nil, defaultError
	}

	h.fileWriter = fileWriter

	return h, nil
}

func (h *sftpHandler) Filecmd(r *sftp.Request) error {
	// it executes only if it is a directoy, before transfer
	if r.Method == "Mkdir" {
		if h.streamDetails.Filename == "" {
			h.streamDetails.Filename = filepath.Base(r.Filepath)
		}
		return nil
	}
	// it executes after transfer
	if r.Method == "Setstat" {
		return nil
	}
	return sftp.ErrSshFxOpUnsupported
}

func (h *sftpHandler) WriteAt(p []byte, off int64) (n int, err error) {
	amount, err := h.fileWriter.Write(p)
	if err != nil {
		return 0, err
	}

	*h.totalSize = *h.totalSize + amount
	if *h.totalSize >= limit {
		fmt.Fprintf(h.stderr, "\n\n%v\n\n", maxLimitError)
		h.server.Close()
		return 0, maxLimitError
	}

	return amount, nil
}
