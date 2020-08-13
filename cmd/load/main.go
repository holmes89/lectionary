package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

const baseDir = "./resources/bible/"

func main() {
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.IsDir() {
			loadBook(f.Name())
		}
	}
}

func loadBook(name string) {
	dat, err := ioutil.ReadFile(baseDir + name)
	if err != nil {
		panic(err)
	}
	version := strings.ReplaceAll(name, ".json", "")
	var bible map[string]interface{}
	if err := json.Unmarshal(dat, &bible); err != nil {
		panic(err)
	}

	for book, chapterMap := range bible {
		for chapter, verseMap := range chapterMap.(map[string]interface{}) {
			for verse, content := range verseMap.(map[string]interface{}) {
				_ = content.(string)
				fmt.Printf("%s\\%s\\%s:%s\n", version, book, chapter, verse)
			}
		}
	}
}
