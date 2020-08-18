package verses

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/holmes89/lectionary/internal"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const (
	verseRegex   = `(\d?\s?\w+)\s?([0-9\-\:\s)]+(?:end)?)?`
	databaseFile = "bible.db"
)

type service struct {
	parser *regexp.Regexp
	db     *bolt.DB
}

func NewService(lc fx.Lifecycle) internal.VerseService {
	parser, err := regexp.Compile(verseRegex)
	if err != nil {
		logrus.WithError(err).Fatal("unable to compile regex")
	}

	logrus.WithField("path", databaseFile).Info("connecting to database")
	conn, err := bolt.Open(databaseFile, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("unable to open database")
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("closing database")
			return conn.Close()
		},
	})

	return &service{
		parser: parser,
		db:     conn,
	}
}

func (s *service) Find(verse string, version internal.Version) ([]internal.Verse, error) {
	matches := s.parser.FindStringSubmatch(verse)
	if len(matches) < 2 {
		return nil, internal.ErrInvalidVerseFormat
	}
	book := matches[1]
	logrus.WithField("book", book).Info("found book")
	if len(matches) > 2 && matches[2] != "" {
		logrus.WithField("verses", matches[2]).Info("searching by verse")
		return s.findVerses(book, matches[2], version)
	}
	return s.findBook(book, version)
}

func (s *service) findVerses(book, verse string, version internal.Version) ([]internal.Verse, error) {
	return nil, nil
}

func (s *service) findBook(book string, version internal.Version) ([]internal.Verse, error) {
	logrus.WithField("book", book).Info("finding by book")
	book = strings.ToLower(book)
	book = strings.TrimSpace(book)
	displayBook := strings.Title(book)
	verses := make([]internal.Verse, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(version)).Bucket([]byte(book))
		b.ForEach(func(k, v []byte) error {
			verse := internal.Verse{
				DisplayName: fmt.Sprintf("%s %s", displayBook, string(k)),
				Content:     string(v),
				Version:     version,
			}
			verses = append(verses, verse)
			return nil
		})
		return nil
	})
	if err != nil {
		logrus.WithError(err).WithField("book", book).Error("unable to find all verses in book")
		return nil, errors.New("failed to fetch verses")
	}
	return verses, nil
}
