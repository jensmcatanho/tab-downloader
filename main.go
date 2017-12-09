package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	url := "https://www.ultimate-guitar.com/tabs/Avantasia_guitar_pro_tabs.htm"

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("tr .tr__lg").Not(".tr__active").Each(func(i int, s *goquery.Selection) {
		if !s.Find("a").HasClass("song js-tp_link") {
			name := s.Find("a").Text()
			link, _ := s.Find("a").Attr("href")
			fmt.Println(name, link)
		}
	})
}
