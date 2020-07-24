package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/widget"
)

/*
	find out video ids by following method
	open webpage e.g. https://www.tiktok.com/@ridhamgayatribhatt

	find appropriate class name
	children = document.getElementsByClassName("jsx-1523213582 video-feed-item-wrapper");

	get list of videos
	list = []; for(var i=0;i<children.length;i++) list.push(children[i].href);
*/

var (
	_jar               = &cookiejar.Jar{}
	_client            = http.DefaultClient
	_proxyURL *url.URL = nil
)

func init() {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	_jar = jar

	_client = &http.Client{
		Jar:       _jar,
		Transport: &http.Transport{},
	}
}

func fill(path string, m map[string]bool) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(strings.NewReader(string(b)))
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for _, record := range records {
		m[record[0]] = true
	}
}

func download(c *http.Client, name, url string) error {
	fp := filepath.Base(name)
	fmt.Println("Downloading ", fp) // This one is getting downloaded
	f := fp + ".mp4"
	if _, err := os.Stat(f); err == nil {
		return nil
	}
	resp, err := c.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(f)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func processWithProxy(p *dialog.ProgressDialog, text, proxy string) error {
	proxy = strings.TrimSpace(proxy)
	if proxy == "" {
		_client.Transport = &http.Transport{}
		_proxyURL = nil
	} else {
		proxyURL, err := url.Parse(proxy)
		if err != nil {
			return err
		}
		_client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
		_proxyURL = proxyURL
	}

	text = strings.TrimSpace(text)
	videoURLs := strings.Split(text, "\n")

	var _finalErr error
	n := len(videoURLs)
	for i, url := range videoURLs {
		p.SetValue(float64(i) / float64(n))
		if err := process(url); err != nil {
			_finalErr = err
		}
	}
	p.SetValue(1.0)

	return _finalErr
}

func process(videoURL string) error {
	var _err error
	fp := filepath.Base(videoURL)
	f := fp + ".mp4"
	if _, err := os.Stat(f); err == nil {
		return errors.New("File already exists")
	}
	req, err := http.NewRequest(http.MethodGet, videoURL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")

	resp, err := _client.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	a := strings.Split(string(b), "\"],\"videoMeta\":")[0]
	splitted := strings.Split(a, "\"urls\":")
	if len(splitted) < 2 {
		return errors.New("Unable to fetch content for video URL")
	}
	a = splitted[1][2:]
	if err = download(_client, videoURL, a); err != nil {
		return err
	}

	return _err
}

func main() {
	app := app.New()

	w := app.NewWindow("Download From Tiktok")

	url := widget.NewMultiLineEntry()
	url.PlaceHolder = "Enter URL(s)"

	proxy := widget.NewEntry()
	proxy.PlaceHolder = "HTTPS proxy"

	submitButton := widget.NewButton("Submit", func() {
		progress := dialog.NewProgress("Downloading", "Downloading", w)
		if err := processWithProxy(progress, url.Text, proxy.Text); err != nil {
			progress.Hide()
			dialog.NewError(err, w).Show()
		} else {
			progress.Hide()
		}

	})

	submitButton.Alignment = widget.ButtonAlignCenter

	screenContent := widget.NewVBox(
		url,
		proxy,
		submitButton,
	)
	w.SetContent(
		screenContent,
	)
	w.Resize(fyne.NewSize(600, 400))
	w.CenterOnScreen()

	w.ShowAndRun()
}
