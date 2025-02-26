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

	"github.com/essentialkaos/translit/v2"
	"github.com/go-chi/chi/v5"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/db"
	"github.com/shemanaev/inpxer/internal/model"
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

	index, err := db.Open(h.cfg.IndexPath, h.cfg.Storage)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	book, err := index.GetBookById(id)
	if err != nil {
		log.Printf("File with id: %s not found in index: %v", id, err)
		notFound(w, id)
		return
	}

	if book.File.IsArchived() {
		data, err := h.getFileFromArchive(book)
		if err != nil {
			notFound(w, id)
			return
		}

		filename := fmt.Sprintf("%s.%s", book.File.Name, book.File.Ext)
		log.Printf("File `%s` for id %s served directly from archive (%s)", filename, id, book.File.Archive)

		addFilenameToHeader(w, book.Title, filename)
		http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(data))
	} else {
		filename, err := h.getDirectFilePath(book)
		if err != nil {
			notFound(w, id)
			return
		}

		log.Printf("File `%s` for id %s served directly from fs", filename, id)

		addFilenameToHeader(w, book.Title, book.File.Name)
		http.ServeFile(w, r, filename)
	}
}

func (h *DownloadHandler) DownloadConverted(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ext := chi.URLParam(r, "ext")

	var converter *config.Converter
	for _, c := range h.cfg.Converters {
		if strings.EqualFold(c.To, ext) {
			converter = c
			break
		}
	}

	if converter == nil {
		log.Printf("Not found converter for `%s`", ext)
		notFound(w, id)
		return
	}

	index, err := db.Open(h.cfg.IndexPath, h.cfg.Storage)
	if err != nil {
		log.Printf("Error opening index: %v", err)
		internalServerError(w)
		return
	}
	defer index.Close()

	book, err := index.GetBookById(id)
	if err != nil {
		log.Printf("File with id: %s not found in index: %v", id, err)
		notFound(w, id)
		return
	}

	if strings.ToLower(book.File.Ext) != converter.From {
		log.Printf("Wrong converter selected for id: %s. expected %s=>%s, got %s=>%s", id, book.File.Ext, ext, converter.From, converter.To)
		notFound(w, id)
		return
	}

	var filename string
	if book.File.IsArchived() {
		data, err := h.getFileFromArchive(book)
		if err != nil {
			notFound(w, id)
			return
		}

		f, err := os.CreateTemp("", "book*."+book.File.Ext)
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

	outDir := os.TempDir()
	outFilename := filepath.Join(outDir, book.LibId+"."+converter.To)
	if strings.Contains(converter.Arguments, "{to_dir}") {
		baseFilename := filepath.Base(filename)
		outFilename = filepath.Join(outDir, strings.TrimSuffix(baseFilename, filepath.Ext(baseFilename))+"."+converter.To)
	}

	args := strings.Replace(converter.Arguments, "{from}", filename, -1)
	args = strings.Replace(args, "{to}", outFilename, -1)
	args = strings.Replace(args, "{to_dir}", outDir, -1)
	cmd := exec.Command(converter.Command, strings.Split(args, " ")...)

	log.Printf("Waiting for converter to finish. %s %s", converter.Command, args)
	if err := cmd.Run(); err != nil {
		log.Printf("Error starting converter: %v", err)
		notFound(w, id)
		return
	}
	defer os.Remove(outFilename)

	log.Printf("Serving converted file from: %s", outFilename)
	addFilenameToHeader(w, book.Title, book.LibId+"."+converter.To)
	http.ServeFile(w, r, outFilename)
}

func (h *DownloadHandler) getDirectFilePath(book *model.Book) (string, error) {
	filename := filepath.Join(h.cfg.LibraryPath, filepath.FromSlash(book.File.Folder), book.File.Name)
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		log.Printf("File `%s` (id: %s) not found: %v", filename, book.LibId, err)
		return "", err
	}

	return filename, nil
}

func (h *DownloadHandler) getFileFromArchive(book *model.Book) ([]byte, error) {
	archivePath := filepath.Join(h.cfg.LibraryPath, book.File.ArchivePath())
	zf, err := zip.OpenReader(archivePath)
	if err != nil {
		log.Printf("Can't open archive `%s` (id: %s) not found: %v", archivePath, book.LibId, err)
		return nil, err
	}
	defer zf.Close()

	bookName := fmt.Sprintf("%s.%s", book.File.Name, book.File.Ext)
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

	log.Printf("File `%s` not found in archive `%s` (id: %s)", bookName, archivePath, book.LibId)
	return nil, os.ErrNotExist
}

func addFilenameToHeader(w http.ResponseWriter, title string, filename string) {
	fileNameTranslit := formatFileNameTranslit(title, filename)
	fileNameUtf8 := formatFileNameUtf8(title, filename)
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s; filename*=UTF-8''%s", fileNameTranslit, fileNameUtf8))
}

func formatFileNameUtf8(title string, filename string) string {
	fileName := fmt.Sprintf("%s-%s", url.PathEscape(title), filename)
	return fileName
}

func formatFileNameTranslit(title string, filename string) string {
	str := translit.ICAO(title)

	replace := map[string]string{
		" ":  "_",
		"/":  "_",
		"\\": "_",
	}
	for s, r := range replace {
		str = strings.ReplaceAll(str, s, r)
	}

	fileName := fmt.Sprintf("%s-%s", str, filename)
	return fileName
}
