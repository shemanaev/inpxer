package indexer

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chelnak/ysmrr"
	"github.com/chelnak/ysmrr/pkg/animations"
	"github.com/chelnak/ysmrr/pkg/colors"
	"github.com/urfave/cli/v2"

	"github.com/shemanaev/inpxer/internal/config"
	"github.com/shemanaev/inpxer/internal/db"
	"github.com/shemanaev/inpxer/internal/model"
	"github.com/shemanaev/inpxer/pkg/inpx"
)

const batchSize = 1000

func Run(cfg *config.MyConfig, filename string, keepDeleted, partial bool) error {
	collection, err := inpx.Open(filename)
	if err != nil {
		log.Printf("Error opening inpx: %s", filename)
		return cli.Exit(err.Error(), 1)
	}
	defer collection.Close()

	if !partial {
		// Delete old index.
		if _, err := os.Stat(cfg.IndexPath); !os.IsNotExist(err) {
			log.Println("Deleting old index...")
			err = os.RemoveAll(cfg.IndexPath)
			if err != nil {
				log.Printf("Error deleting old index: %s", cfg.IndexPath)
				return cli.Exit(err.Error(), 1)
			}
		}
	}

	idx, err := db.Create(cfg.IndexPath, cfg.Language)
	if err != nil {
		log.Printf("Error opening or creating index: %s", cfg.IndexPath)
		return cli.Exit(err.Error(), 1)
	}
	defer idx.Close()

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	s := sm.AddSpinner("Indexing...")
	sm.Start()
	defer sm.Stop()

	start := time.Now()

	var recordsCount, deletedCount int
	duplicates := make(map[int]int)
	books := make([]*model.Book, 0)
	for book := range collection.Stream() {
		recordsCount = recordsCount + 1
		if book.Deleted && !keepDeleted {
			deletedCount = deletedCount + 1
			continue
		}

		_, exist := duplicates[book.LibId]
		if exist {
			duplicates[book.LibId] += 1
		} else {
			duplicates[book.LibId] = 1
			books = append(books, model.NewBook(book))
		}

		if len(books) > batchSize {
			err := idx.AddBooks(books, partial)
			if err != nil {
				s.Error()
				return cli.Exit(err.Error(), 1)
			}
			s.UpdateMessage(fmt.Sprintf("Processed: %d", recordsCount))
			books = make([]*model.Book, 0)
		}
	}

	if len(books) > 0 {
		err := idx.AddBooks(books, partial)
		if err != nil {
			s.Error()
			return cli.Exit(err.Error(), 1)
		}
		s.UpdateMessage(fmt.Sprintf("Processed: %d", recordsCount))
	}

	duplicatesCount := 0
	for _, v := range duplicates {
		if v > 1 {
			duplicatesCount = duplicatesCount + v - 1
		}
	}

	if err := collection.Err(); err != nil {
		s.Error()
		log.Printf("Error parsing inpx file: %s", filename)
		return cli.Exit(err.Error(), 1)
	}

	s.UpdateMessage("Done")
	s.Complete()
	elapsed := time.Since(start)
	log.Printf("Processed: %d, imported: %d, duplicates: %d, deleted: %d. (Took %s)", recordsCount, recordsCount-duplicatesCount-deletedCount, duplicatesCount, deletedCount, elapsed)

	return nil
}
