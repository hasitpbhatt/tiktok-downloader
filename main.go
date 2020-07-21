package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	_jar    = &cookiejar.Jar{}
	_videos = map[string]bool{}
)

func init() {
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		panic(err)
	}
	_jar = jar
	fill("list.txt", _videos)
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
	fmt.Println(fp) // This one is getting downloaded
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

func process(c *http.Client, url string) {
	fp := filepath.Base(url)
	f := fp + ".mp4"
	if _, err := os.Stat(f); err == nil {
		return
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36")

	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	// if err := ioutil.WriteFile("z.html", b, 0777); err != nil {
	// 	panic(err)
	// }
	a := strings.Split(string(b), "\"],\"videoMeta\":")[0]
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(url)
		}
	}()
	a = strings.Split(a, "\"urls\":")[1][2:]
	if err = download(c, url, a); err != nil {
		panic(err)
	}
}

func main() {
	// proxy := "http://52.179.18.244:8080"
	proxy := "http://137.220.34.109:8080"
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}
	c := &http.Client{
		Jar: _jar,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	for i := range _videos {
		process(c, i)
		break
	}
}
