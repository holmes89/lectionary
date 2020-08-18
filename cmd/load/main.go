package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/boltdb/bolt"

	"github.com/sirupsen/logrus"
)

const baseDir = "./resources/bible/"

func main() {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	db, err := bolt.Open("bible.db", 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create bolt file")
	}
	defer db.Close()
	for _, f := range files {
		if !f.IsDir() {
			loadBook(f.Name(), db)
		}
	}

	db.View(func(tx *bolt.Tx) error {
		res := tx.Bucket([]byte("nlt")).Bucket([]byte("john")).Get([]byte("3:17"))
		fmt.Println(string(res))
		return nil
	})
}

func loadBook(name string, db *bolt.DB) {
	dat, err := ioutil.ReadFile(baseDir + name)
	if err != nil {
		panic(err)
	}
	version := strings.ReplaceAll(name, ".json", "")
	var bible map[string]interface{}
	if err := json.Unmarshal(dat, &bible); err != nil {
		panic(err)
	}
	tx, err := db.Begin(true)
	if err != nil {
		logrus.WithError(err).Fatal("unable to create transaction")
	}
	vbucket, err := tx.CreateBucketIfNotExists([]byte(version))
	if err != nil {
		logrus.WithError(err).WithField("version", version).Fatal("unable to create version bucket")
	}
	for book, chapterMap := range bible {
		bbucket, err := vbucket.CreateBucketIfNotExists([]byte(book))
		if err != nil {
			logrus.WithError(err).WithField("book", book).Fatal("unable to create book bucket")
		}
		for chapter, verseMap := range chapterMap.(map[string]interface{}) {
			for verse, content := range verseMap.(map[string]interface{}) {
				contentText := content.(string)
				key := fmt.Sprintf("%s:%s", chapter, verse)
				bbucket.Put([]byte(key), []byte(contentText))
			}
		}
	}
	if err := tx.Commit(); err != nil {
		logrus.WithError(err).Fatal("unable to commit transaction")
	}
}
