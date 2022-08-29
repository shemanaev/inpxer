package server

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/model"
	"github.com/shemanaev/inpxer/internal/storage"
)

type DownloadHandler struct {
	cfg *config.MyConfig
}

func NewDownloadHandler(cfg *config.MyConfig) *DownloadHandler {
	return &DownloadHandler{
		cfg: cfg,
	}
}

func (h *DownloadHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	index, err := storage.Open(h.cfg.IndexPath, h.cfg.Language, false)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	book, err := index.FindById(id)
	if err != nil {
		log.Printf("File with id: %s not found in index: %v", id, err)
		notFound(w, id)
		return
	}

	if book.Folder == "" {
		data, err := h.getFileFromArchive(book)
		if err != nil {
			notFound(w, id)
			return
		}

		filename := fmt.Sprintf("%s.%s", book.File, book.Ext)
		log.Printf("File `%s` for id %s served directly from archive (%s)", filename, id, book.Archive)

		addFilenameToHeader(w, book.Title, filename)
		http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(data))
	} else {
		filename, err := h.getDirectFilePath(book)
		if err != nil {
			notFound(w, id)
			return
		}

		log.Printf("File `%s` for id %s served directly from fs", filename, id)

		addFilenameToHeader(w, book.Title, book.File)
		http.ServeFile(w, r, filename)
	}
}

func (h *DownloadHandler) DownloadConverted(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ext := strings.ToLower(chi.URLParam(r, "ext"))

	var converter *config.Converter
	for _, c := range h.cfg.Converters {
		if strings.ToLower(c.To) == ext {
			converter = c
			break
		}
	}

	if converter == nil {
		log.Printf("Not found converter for `%s`", ext)
		notFound(w, id)
		return
	}

	index, err := storage.Open(h.cfg.IndexPath, h.cfg.Language, false)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	book, err := index.FindById(id)
	if err != nil {
		log.Printf("File with id: %s not found in index: %v", id, err)
		notFound(w, id)
		return
	}

	if strings.ToLower(book.Ext) != converter.From {
		log.Printf("Wrong converter selected for id: %s. expected %s=>%s, got %s=>%s", id, book.Ext, ext, converter.From, converter.To)
		notFound(w, id)
		return
	}

	var filename string
	if book.Folder == "" {
		data, err := h.getFileFromArchive(book)
		if err != nil {
			notFound(w, id)
			return
		}

		f, err := os.CreateTemp("", "book*."+book.Ext)
		if err != nil {
			log.Printf("Error creating temp file: %v", err)
			notFound(w, id)
			return
		}

		if _, err := f.Write(data); err != nil {
			log.Printf("Error writing to temp file: %v", err)
			notFound(w, id)
			return
		}

		if err := f.Close(); err != nil {
			log.Printf("Error closing temp file: %v", err)
			notFound(w, id)
			return
		}

		filename = f.Name()
		defer os.Remove(filename)
	} else {
		filename, err = h.getDirectFilePath(book)
		if err != nil {
			notFound(w, id)
			return
		}
	}

	outFilename := filepath.Join(os.TempDir(), book.LibId+"."+converter.To)
	args := strings.Replace(converter.Arguments, "{from}", filename, -1)
	args = strings.Replace(args, "{to}", outFilename, -1)
	cmd := exec.Command(converter.Command, strings.Split(args, " ")...)

	log.Printf("Waiting for converter to finish. %s %s", converter.Command, args)
	if err := cmd.Run(); err != nil {
		log.Printf("Error starting converter: %v", err)
		notFound(w, id)
		return
	}
	defer os.Remove(outFilename)

	addFilenameToHeader(w, book.Title, book.LibId+"."+converter.To)
	http.ServeFile(w, r, outFilename)
}

func (h *DownloadHandler) getDirectFilePath(book *model.Book) (string, error) {
	filename := filepath.Join(h.cfg.LibraryPath, filepath.FromSlash(book.Folder), book.File)
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		log.Printf("File `%s` (id: %s) not found: %v", filename, book.LibId, err)
		return "", err
	}

	return filename, nil
}

func (h *DownloadHandler) getFileFromArchive(book *model.Book) ([]byte, error) {
	archivePath := filepath.Join(h.cfg.LibraryPath, book.Archive+".zip")
	zf, err := zip.OpenReader(archivePath)
	if err != nil {
		log.Printf("Can't open archive `%s` (id: %s) not found: %v", archivePath, book.LibId, err)
		return nil, err
	}
	defer zf.Close()

	bookName := fmt.Sprintf("%s.%s", book.File, book.Ext)
	for _, file := range zf.File {
		if file.Name == bookName {
			content, err := file.Open()
			if err != nil {
				log.Printf("Can't open file `%s` in archive `%s` (id: %s) not found: %v", bookName, archivePath, book.LibId, err)
				return nil, err
			}

			data, err := io.ReadAll(content)
			if err != nil {
				log.Printf("Can't read file `%s` in archive `%s` (id: %s) not found: %v", bookName, archivePath, book.LibId, err)
				content.Close()
				return nil, err
			}
			content.Close()

			return data, nil
		}
	}

	return nil, os.ErrNotExist
}

func addFilenameToHeader(w http.ResponseWriter, title string, filename string) {
	fileName := fmt.Sprintf("%s-%s", url.PathEscape(title), filename)
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
}
