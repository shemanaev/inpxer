package inpx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFlibustaFb2Local(t *testing.T) {
	assert := assert.New(t)
	collection, err := Open("testdata/flibusta_fb2_local.inpx")
	if err != nil {
		t.Fatalf(`collection is not opened: %v`, err)
	}
	defer collection.Close()

	assert.Equal("Flibusta FB2 Local", collection.Name)
	assert.Equal(65536, collection.Id)
	assert.Equal("Локальная коллекция библиотеки Флибуста (только FB2)", collection.Comment)

	assert.Equal("20220601", collection.Version)

	bookCount := 0
	for book := range collection.Stream() {
		bookCount = bookCount + 1

		if book.LibId == 166370 {
			assert.Equal("Болельщик. С. О'Нэн,  С. Кинг. Рецензия", book.Title)
			assert.Equal("Рецензии", book.Series)
			assert.Equal(time.Date(2009, 9, 20, 0, 0, 0, 0, time.UTC), book.PublishedDate)
			assert.Equal([]string{"nonf_criticism"}, book.Genres)
			assert.Equal("166370", book.File.Name)
			assert.Equal("fb2", book.File.Ext)
			assert.Equal(9236, book.File.Size)
			assert.Equal("fb2-166043-168102", book.File.Archive)
			assert.Equal("ru", book.Language)
			authors := []Author{
				{
					LastName:   "Кинг",
					FirstName:  "Стивен",
					MiddleName: "",
				},
				{
					LastName:   "Вебер",
					FirstName:  "Виктор",
					MiddleName: "Анатольевич",
				},
				{
					LastName:   "О'Нэн",
					FirstName:  "Стюарт",
					MiddleName: "",
				},
			}
			assert.Equal(authors, book.Authors)
		}
	}

	assert.Equal(50, bookCount)
}

func TestFlibustaRev20(t *testing.T) {
	assert := assert.New(t)
	collection, err := Open("testdata/flibusta.all-rev2.0-2022-07-04.inpx")
	if err != nil {
		t.Fatalf(`collection is not opened: %v`, err)
	}
	defer collection.Close()

	assert.Equal("Flibusta.ALL rev2.0 2007 - 2022 (July 4)", collection.Name)
	assert.Equal(65537, collection.Id)
	assert.Equal("Total: 457789 fb2 + 78628 usr books", collection.Comment)

	assert.Equal("20220704", collection.Version)

	bookCount := 0
	for book := range collection.Stream() {
		bookCount = bookCount + 1

		if book.LibId == 486580 {
			assert.Equal("Беседы о физике и технике", book.Title)
			assert.Equal("", book.Series)
			assert.Equal(time.Date(2017, 5, 13, 0, 0, 0, 0, time.UTC), book.PublishedDate)
			assert.Equal([]string{"sci_tech", "science", "tbg_secondary"}, book.Genres)
			assert.Equal("486580.fb2.zip", book.File.Name)
			assert.Equal("fb2", book.File.Ext)
			assert.Equal(1477459, book.File.Size)
			assert.Equal("2017\\05\\13\\", book.File.Folder)
			assert.Equal("ru", book.Language)
			authors := []Author{
				{
					LastName:   "Глухов",
					FirstName:  "Николай",
					MiddleName: "Данилович",
				},
				{
					LastName:   "Камышанченко",
					FirstName:  "Николай",
					MiddleName: "Васильевич",
				},
				{
					LastName:   "Самойленко",
					FirstName:  "Петр",
					MiddleName: "Иванович",
				},
			}
			assert.Equal(authors, book.Authors)
		}
	}

	assert.Equal(53, bookCount)
}

func TestFlibustaAllLocal20250202(t *testing.T) {
	assert := assert.New(t)
	collection, err := Open("testdata/flibusta_all_local-2025-02-02.inpx")
	if err != nil {
		t.Fatalf(`collection is not opened: %v`, err)
	}
	defer collection.Close()

	assert.Equal("Flibusta Offline 2 February 2025", collection.Name)
	assert.Equal(65537, collection.Id)
	assert.Equal("Flibusta. A local collection. Total: 771171 books", collection.Comment)

	assert.Equal("20250202", collection.Version)

	bookCount := 0
	for book := range collection.Stream() {
		bookCount = bookCount + 1

		if book.LibId == 754865 {
			assert.Equal("Книги фантастики українських видавництв за 1960-1991 рр.", book.Title)
			assert.Equal("", book.Series)
			assert.Equal(time.Date(2023, 11, 01, 0, 0, 0, 0, time.UTC), book.PublishedDate)
			assert.Equal([]string{"reference", "sf_etc"}, book.Genres)
			assert.Equal("754865", book.File.Name)
			assert.Equal("rtf", book.File.Ext)
			assert.Equal(1436043, book.File.Size)
			assert.Equal("f.usr-754754-759835.zip", book.File.Folder)
			assert.Equal("uk", book.Language)
			authors := []Author{
				{
					LastName:   "Левченко",
					FirstName:  "Александр",
					MiddleName: "Николаевич",
				},
			}
			assert.Equal(authors, book.Authors)
		}
		if book.LibId == 362143 {
			assert.Equal("Сан Феличе", book.Title)
			assert.Equal("", book.Series)
			assert.Equal(time.Date(2014, 05, 03, 0, 0, 0, 0, time.UTC), book.PublishedDate)
			assert.Equal([]string{"prose_history", "adventure", "sci_culture", "design"}, book.Genres)
			assert.Equal("362143", book.File.Name)
			assert.Equal("fb2", book.File.Ext)
			assert.Equal(2607193, book.File.Size)
			assert.Equal("f.fb2-361575-365133.zip", book.File.Folder)
			assert.Equal("bg", book.Language)
			authors := []Author{
				{
					LastName:   "Дюма",
					FirstName:  "Александр",
					MiddleName: "",
				},
			}
			assert.Equal(authors, book.Authors)
		}
	}

	assert.Equal(3, bookCount)
}
