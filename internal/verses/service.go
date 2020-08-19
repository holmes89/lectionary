package verses

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
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
	verse = strings.TrimSpace(verse)
	matches := s.parser.FindStringSubmatch(verse)
	if len(matches) < 2 {
		return nil, internal.ErrInvalidVerseFormat
	}
	book := matches[1]
	book = strings.ToLower(book)

	logrus.WithField("book", book).Info("found book")
	if len(matches) > 2 && matches[2] != "" {
		logrus.WithField("verses", matches[2]).Info("searching by verse")
		verses, err := s.parseVerse(book, matches[2], version)
		if err != nil {
			return nil, internal.ErrInvalidVerseFormat
		}
		return s.findVerses(book, verses, version)
	}
	return s.findBook(book, version)
}

func (s *service) parseVerse(book, verse string, version internal.Version) ([]string, error) {
	var verses []string
	logrus.WithField("verse", verse).Info("parsing verses")
	// 1 -> 1:1-end +
	// 1:1 +
	// 1:1-3:13 -> 1:1-end 2:1-end 3:1-13 +
	// 1:1-end -> 1:1-53 -> [1:1, 1:2, ...] +
	// 1:1-2 4-5 -> [1:1,1:2,1:4,1:5] +
	// 1:1-2 -> [1:1,1:2] +
	// 1:1-3 2:3-4 -> [1:1,1:2,2:3,2:4] +
	var chapter string
	for _, v := range strings.Split(verse, " ") { // case 1:1-3 2:3-4 -> [1:1,1:2,2:3,2:4]
		if !strings.Contains(v, ":") && chapter == "" { // case 1 -> 1:1-end
			v = v + ":1-end"
		}
		if !strings.Contains(v, ":") { // case 1:1-2 4-5 -> 1:1-2 1:4-5
			v = chapter + ":" + v
		}
		// So at this point we can ensure that the values are either 1:1-end or 1:1-3
		sv := strings.Split(v, ":") // separate chapter and verse
		chapter = sv[0]
		v = sv[1]

		if len(sv) == 3 { // 1:1-3:13 -> 1:1-end 2:1-end 3:1-13, it was split three ways
			// sv = [1, 1-3, 13] ->
			sv2 := strings.Split(sv[1], "-")

			cs, _ := strconv.Atoi(sv[0]) //this seems dangerous, maybe cleanup
			vs, _ := strconv.Atoi(sv2[0])
			ce, _ := strconv.Atoi(sv2[1])
			ve, _ := strconv.Atoi(sv[2])

			v = fmt.Sprintf("%d:%d-end ", cs, vs) //start chapter and verses
			for i := cs + 1; i < ce; i++ {        // middle verses
				v = fmt.Sprintf("%s%d:1-end ", v, i)
			}
			v = fmt.Sprintf("%s%d:1-%d", v, ce, ve) // end chapter and verses
			res, err := s.parseVerse(book, v, version)
			if err != nil {
				return nil, err
			}
			verses = append(verses, res...)
			continue
		}

		if strings.Contains(v, "end") { // case 1:1-end -> 1:1-53
			end, err := s.getEnd(book, chapter, version)
			if err != nil {
				logrus.WithError(err).Error("unable to find last verse")
				return nil, errors.New("unable to find verses")
			}
			v = strings.ReplaceAll(v, "end", strconv.Itoa(end))
		}

		if strings.Contains(v, "-") { // 1:1-2 -> [1:1,1:2]
			logrus.WithField("verse", v).Info("finding range")
			verseRange := strings.Split(v, "-")
			vs, _ := strconv.Atoi(verseRange[0])
			ve, _ := strconv.Atoi(verseRange[1])
			for i := vs; i <= ve; i++ {
				verses = append(verses, fmt.Sprintf("%s:%d", chapter, i))
			}
		} else { // we can only assume that this is 1:1 ?
			verses = append(verses, chapter+":"+v)
		}
	}
	return verses, nil
}

func (s *service) getEnd(book, chapter string, version internal.Version) (int, error) {
	logrus.WithFields(logrus.Fields{
		"chapter": chapter,
		"book":    book,
		"version": version,
	}).Info("finding verse count")
	var result int
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(version)).Bucket([]byte("versecount"))
		key := fmt.Sprintf("%s:%s", book, chapter)
		v := bucket.Get([]byte(key))
		if v == nil {
			return errors.New("unable to find book and chapter")
		}
		result, _ = strconv.Atoi(string(v))
		return nil
	})
	return result, err
}

func (s *service) findVerses(book string, verses []string, version internal.Version) ([]internal.Verse, error) {
	logrus.WithFields(logrus.Fields{
		"book":       book,
		"versecount": len(verses),
	}).Info("finding by verses")
	displayBook := strings.Title(book)
	res := make([]internal.Verse, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(version)).Bucket([]byte(book))
		for _, k := range verses {
			content := b.Get([]byte(k))
			verse := internal.Verse{
				DisplayName: fmt.Sprintf("%s %s", displayBook, string(k)),
				Content:     string(content),
				Version:     version,
			}
			res = append(res, verse)
		}
		return nil
	})
	if err != nil {
		logrus.WithError(err).WithField("book", book).Error("unable to find all verses in book")
		return nil, errors.New("failed to fetch verses")
	}
	return res, nil
}

func (s *service) findBook(book string, version internal.Version) ([]internal.Verse, error) {
	logrus.WithField("book", book).Info("finding by book")
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
