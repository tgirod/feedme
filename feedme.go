package main

import (
	"encoding/json"
	"fmt"
	fp "github.com/iand/feedparser"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type SourceList struct {
	Filename string
	Sources  []Source
}

type Source struct {
	Title  string
	Url    string
	Latest time.Time
}

type Link struct {
	Url   string
	Title string
	When  time.Time
}

func NewSourceList(filename string) *SourceList {
	sl := SourceList{Filename: filename}
	if err := sl.Load(); err != nil {
		log.Fatal(err)
	}
	return &sl
}

func (sl *SourceList) Load() error {
	file, err := os.Open(sl.Filename)
	defer file.Close()
	if err != nil {
		return err
	}
	dec := json.NewDecoder(file)
	for {
		var s Source
		err := dec.Decode(&s)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		sl.Sources = append(sl.Sources, s)
	}
	return nil
}

func (sl *SourceList) Save() error {
	file, err := os.OpenFile(sl.Filename, os.O_RDWR|os.O_TRUNC, os.ModePerm)
	defer file.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(file)
	for _, s := range sl.Sources {
		err := enc.Encode(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sl *SourceList) AddSource(url string) {
	s := Source{Url: url}
	for _, s = range sl.Sources {
		if s.Url == url {
			return
		}
	}
	sl.Sources = append(sl.Sources, s)
}

func (sl *SourceList) Fetch() (map[Source][]Link, error) {
	links := map[Source][]Link{}
	for _, s := range sl.Sources {
		l, err := s.Fetch()
		if err != nil {
			return nil, err
		}
		links[s] = l
	}
	return links, nil
}

func (s *Source) Fetch() ([]Link, error) {
	// grab feed
	resp, err := http.Get(s.Url)
	if err != nil {
		return nil, err
	}
	// parse feed
	feed, err := fp.NewFeed(resp.Body)
	if err != nil {
		return nil, err
	}
	// get feed title
	s.Title = feed.Title
	// get new items
	links := []Link{}
	for _, it := range feed.Items {
		if !it.When.After(s.Latest) {
			break
		}
		links = append(links, Link{it.Link, it.Title, it.When})
	}
	// change latest fetch
	if len(links) > 0 {
		s.Latest = links[0].When
	}
	return links, nil
}

func main() {
	sl := NewSourceList("feedlist.json")
	sl.AddSource("http://www.questionablecontent.net/QCRSS.xml")
	items, err := sl.Sources[0].Fetch()
	if err != nil {
		log.Fatal(err)
	}
	for _, i := range items {
		fmt.Println(i)
	}
	sl.Save()
}
