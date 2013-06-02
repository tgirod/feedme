package main

import (
	"encoding/json"
	"flag"
	"fmt"
	fp "github.com/iand/feedparser"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var configFile = flag.String("config", os.ExpandEnv("${HOME}/.feedme"), "config file to use. default is $HOME/.feedme")

type Source struct {
	Url    string
	Latest string
}

type SourceList []*Source

func (sl *SourceList) Load(filename string) {
	// open file
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	defer file.Close()
	if err != nil {
		log.Println(err)
		return
	}
	// load json content
	dec := json.NewDecoder(file)
	for {
		var source Source
		err := dec.Decode(&source)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			return
		}
		(*sl) = append((*sl), &source)
	}
}

func (sl *SourceList) Save(filename string) {
	// open file
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	defer file.Close()
	if err != nil {
		fmt.Println(err)
	}
	enc := json.NewEncoder(file)
	for _, feed := range *sl {
		err := enc.Encode(feed)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (sl *SourceList) AddSource(url string) {
	for _, f := range *sl {
		if f.Url == url {
			return
		}
	}
	(*sl) = append((*sl), &Source{Url: url})
}

func (sl *SourceList) DeleteSource(url string) {
	for i, s := range *sl {
		if s.Url == url {
			*sl = append((*sl)[:i], (*sl)[i+1:]...)
		}
	}
}

func (sl *SourceList) Fetch() {
	// launch parallel fetch
	c := make(chan *fp.Feed)
	for _, source := range *sl {
		go source.Fetch(c)
	}
	// print results as they arrive
	nb := 0
	for _, _ = range *sl {
		f := <-c
		if f != nil {
			nb++
			PrintFeed(f)
		}
	}
	log.Printf("parsed %d feeds, %d with new contents\n", len(*sl), nb)
}

func (s *Source) Fetch(c chan *fp.Feed) {
	// grab feed
	resp, err := http.Get(s.Url)
	if err != nil {
		log.Println(err)
		c <- nil
		return
	}
	// parse feed
	feed, err := fp.NewFeed(resp.Body)
	if err != nil {
		log.Println(err)
		c <- nil
		return
	}
	// drop the old items
	items := []*fp.FeedItem{}
	for _, item := range feed.Items {
		if item.Link != s.Latest {
			items = append(items, item)
		} else {
			break
		}
	}
	feed.Items = items
	// update Latest item seen
	if len(feed.Items) > 0 {
		s.Latest = feed.Items[0].Link
	}
	// return result
	if len(items) == 0 {
		c <- nil
		return
	} else {
		c <- feed
	}
}

func PrintFeed(f *fp.Feed) {
	title := strings.TrimSpace(f.Title)
	if title == "" {
		title = "[nameless feed]"
	}
	fmt.Printf("\n%s\n", title)
	for _, i := range f.Items {
		fmt.Printf(" - [%s](%s)\n", i.Title, i.Link)
	}
}

func init() {
	flag.Parse()
	if flag.NArg() == 0 {
		os.Exit(0)
	}
}

func main() {
	sl := SourceList{}
	sl.Load(*configFile)

	if flag.NArg() == 0 {
		os.Exit(0)
	}

	switch flag.Arg(0) {
	case "add":
		for _, source := range flag.Args()[1:] {
			sl.AddSource(source)
		}
		sl.Save(*configFile)
	case "delete":
		for _, source := range flag.Args()[1:] {
			sl.DeleteSource(source)
		}
		sl.Save(*configFile)
	case "fetch":
		sl.Fetch()
		sl.Save(*configFile)
	case "list":
		for _, source := range sl {
			fmt.Println(source.Url)
		}
	default:
		os.Exit(0)
	}

}
