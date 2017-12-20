package representations

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Tab struct {
	Name    string
	Band    string
	ID      string
	Referer string
}

func NewTab(name, band, id, referer string) *Tab {
	return &Tab{
		Name:    name,
		Band:    band,
		ID:      id,
		Referer: referer,
	}
}

func (t *Tab) Download() (err error) {
	matches, _ := filepath.Glob(fmt.Sprintf("bands/%v/%v", t.Band, t.Name) + ".*")
	for _, file := range matches {
		if _, err = os.Stat(file); err == nil {
			log.Printf("%v file already exists. Fast-forwarding...", t.Name)
			return
		} else {
			return
		}
	}


	
	client := http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://tabs.ultimate-guitar.com/tab/download?id=%v", t.ID), nil)
	req.Header.Add("Referer", t.Referer)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	splat := strings.Split(resp.Header.Get("Content-Disposition"), ".")
	extension := splat[len(splat)-1]
	out, err := os.Create(fmt.Sprintf("bands/%v/%v.%v", t.Band, t.Name, extension))
	if err != nil {
		return
	}
	defer out.Close()

	log.Printf("Downloading %v file (#%v)...", t.Name, t.ID)
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	return
}
