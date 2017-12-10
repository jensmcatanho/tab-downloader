package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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
			tab := s.Find("a")
			name := tab.Text()
			log.Printf("Processing %v tab...", name)
			tabURL, _ := tab.Attr("href")

			doc, err := goquery.NewDocument(tabURL)
			if err != nil {
				log.Fatal(err)
			}

			doc.Find("div").Each(func(i int, s *goquery.Selection) {
				if s.HasClass("textversbox") {
					tabID, _ := s.Find("input").Attr("value")
					downloadFile(name, tabID, tabURL)
				}
			})
		}
	})
}

func downloadFile(filename, id, referer string) {
	if _, err := os.Stat(fmt.Sprint("avantasia/", filename)); err == nil {
		log.Println("File already exists. Fast-forwarding...")
		return
	}

	out, err := os.Create(fmt.Sprint("avantasia/", filename))
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://tabs.ultimate-guitar.com/tab/download?id=%v", id), nil)
	req.Header.Add("Host", "tabs.ultimate-guitar.com")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:57.0) Gecko/20100101 Firefox/57.0")
	req.Header.Add("Accept-Language", "pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Referer", referer)
	req.Header.Add("Connection", "keep-alive")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}
