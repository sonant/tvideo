package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	logger    = logrus.New()
	transport = &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}
)

// Track struct for pasing url of video
type Track struct {
	Track struct {
		PlaybackURL string `json:"playbackUrl"`
	} `json:"track"`
}

func converter(id string) (string, error) {
	_ = os.Remove("./videos/" + id + ".mkv")
	time.Sleep(time.Millisecond * 1000)
	// Get bearer
	resp, err := client.Get("https://twitter.com/i/videos/tweet/" + id)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	//body, err := gzip.NewReader(resp.Body)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	re, _ := regexp.Compile("src=\"(.*)\"")
	var bearer string
	if url := re.FindSubmatch(b); url != nil {
		re, _ := regexp.Compile(`(?m)authorization:\"Bearer (.*)\",\"x-csrf`)
		resp, err := client.Get(string(url[1]))
		if err != nil {
			return "", err
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		if r := re.FindSubmatch(b); r != nil {
			bearer = string(r[1])
		}
		resp.Body.Close()
	}
	logger.Infoln(bearer)

	// // Not need
	// // Get Activation
	// time.Sleep(time.Millisecond * 1000)
	// url, _ := url.Parse("https://api.twitter.com/1.1/guest/activate.json")
	// request := &http.Request{
	// 	Method: "POST",
	// 	URL:    url,
	// 	Header: http.Header{"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}, "authorization": []string{"Bearer " + bearer}},
	// }
	// resp, err = client.Do(request)
	// if err != nil {
	// 	return "", err
	// }

	// Get video parameters
	time.Sleep(time.Millisecond * 1000)
	url, _ := url.Parse("https://api.twitter.com/1.1/videos/tweet/config/" + id + ".json")
	request := &http.Request{
		Method: "GET",
		URL:    url,
		Header: http.Header{"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}, "origin": []string{"https://twitter.com"}, "x-guest-token": []string{"1098316853645111296"}, "referer": []string{"https://twitter.com/i/videos/tweet/" + id}, "authorization": []string{"Bearer " + bearer}},
	}
	resp, err = client.Do(request)
	if err != nil {
		return "", err
	}
	logger.Infoln("x-rate-limit-limit:", resp.Header.Get("x-rate-limit-limit"))
	logger.Infoln("x-rate-limit-remaining:", resp.Header.Get("x-rate-limit-remaining"))
	logger.Infoln("x-rate-limit-reset:", resp.Header.Get("x-rate-limit-reset"))
	body, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	var videoURL Track
	// logger.Infof("%s", b)
	if err := json.Unmarshal(b, &videoURL); err != nil {
		return "", err
	}
	time.Sleep(time.Millisecond * 1000)
	videDescription, err := client.Get(videoURL.Track.PlaybackURL)
	if err != nil {
		return "", err
	}
	// b, err = ioutil.ReadAll(videDescription.Body)
	// if err != nil {
	// 	logger.Fatalln(err)
	// }
	time.Sleep(time.Millisecond * 1000)
	scanner := bufio.NewScanner(videDescription.Body)
	for scanner.Scan() {
		t := scanner.Text()
		if []byte(t)[0] == '/' {
			convert := exec.Command("ffmpeg", "-i", "https://video.twimg.com"+t, "-c", "copy", "./videos/"+id+".mkv")
			convert.Stdout = os.Stdout
			convert.Stderr = os.Stderr
			if convert.Run() != nil {
				videDescription.Body.Close()
				return "", err
			}
			break
		}
	}
	videDescription.Body.Close()
	return id, nil
}