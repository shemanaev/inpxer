package server

import (
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/urfave/cli/v2"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/i18n"
	"github.com/shemanaev/inpxer/ui"
)

const PageSize = 10

func init() {
	_ = mime.AddExtensionType(".azw", "application/vnd.amazon.ebook")
	_ = mime.AddExtensionType(".azw3", "application/vnd.amazon.ebook")
	_ = mime.AddExtensionType(".mobi", "application/x-mobipocket-ebook")
	_ = mime.AddExtensionType(".epub", "application/epub+zip")
	_ = mime.AddExtensionType(".fb2", "text/fb2+xml")
	_ = mime.AddExtensionType(".fb2.zip", "application/fb2+zip")
	_ = mime.AddExtensionType(".cbz", "application/x-cbz")
	_ = mime.AddExtensionType(".cbr", "application/x-cbr")
	_ = mime.AddExtensionType(".djv", "image/x-djvu")
	_ = mime.AddExtensionType(".djvu", "image/x-djvu")
}

func Run(cfg *config.MyConfig) error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.CleanPath)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Compress(5))

	t, err := i18n.GetLocalizer(cfg.Language)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	fs := http.FileServer(http.FS(ui.StaticFiles))
	r.Handle("/static/*", fs)

	web := NewWebHandler(cfg, t)
	r.Get("/", web.Home)
	r.Get("/search", web.Search)

	download := NewDownloadHandler(cfg)
	r.Route("/download", func(r chi.Router) {
		r.Get("/{id}", download.Download)
		r.Get("/{id}/{ext}", download.DownloadConverted)
	})

	opds := NewOpdsHandler(cfg, t)
	r.Get("/opensearch.xml", opds.OpenSearchDescription)
	r.Route("/opds", func(r chi.Router) {
		r.Get("/", opds.Root)
		r.Get("/search", opds.Search)
	})

	err = http.ListenAndServe(cfg.Listen, r)
	if err != nil {
		return cli.Exit(err.Error(), 1)
	}

	return nil
}
