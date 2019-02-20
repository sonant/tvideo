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

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
}

// Track struct for pasing url of video
type Track struct {
	Track struct {
		PlaybackURL string `json:"playbackUrl"`
	} `json:"track"`
}

func main() {
	// url, _ := url.Parse("https://twitter.com/i/videos/1096941411180707840") //"https://api.twitter.com/1.1/videos/tweet/config/1096941411180707840.json")
	// request := &http.Request{
	// 	Method: "OPTIONS",
	// 	URL:    url,
	// 	Header: http.Header{"access-control-request-headers": []string{"authorization", "x-csrf-token"}, "access-control-request-method": []string{"POST"}, "user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}},
	// 	//Header: http.Header{"access-control-request-headers": []string{"authorization", "x-csrf-token", "x-guest-token"}, "access-control-request-method": []string{"GET"}, "user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}},
	// }
	// resp, err := client.Do(request) //"https://twitter.com/i/videos/1096941411180707840")
	// if err != nil {
	// 	logger.Errorln(err)
	// 	os.Exit(1)
	// }

	_ = os.Remove("out.mp4")

	// Get bearer
	resp, err := client.Get("https://twitter.com/i/videos/1096941411180707840")
	if err != nil {
		logger.Fatalln("Get bearer ", err)
	}
	defer resp.Body.Close()
	//body, err := gzip.NewReader(resp.Body)
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Fatalln(err)
	}
	re, _ := regexp.Compile("src=\"(.*)\"")
	var bearer string
	if url := re.FindSubmatch(b); url != nil {
		re, _ := regexp.Compile(`(?m)authorization:\"Bearer (.*)\",\"x-csrf`)
		resp, err := client.Get(string(url[1]))
		if err != nil {
			logger.Fatalln(err)
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Fatalln(err)
		}
		if r := re.FindSubmatch(b); r != nil {
			bearer = string(r[1])
		}
		resp.Body.Close()
	}

	// Not need
	// Get Activation
	time.Sleep(time.Millisecond * 500)
	url, _ := url.Parse("https://api.twitter.com/1.1/guest/activate.json")
	request := &http.Request{
		Method: "POST",
		URL:    url,
		Header: http.Header{"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}, "authorization": []string{"Bearer " + bearer}},
	}
	resp, err = client.Do(request)
	if err != nil {
		logger.Errorln(err)
		os.Exit(1)
	}

	// Get video parameters
	time.Sleep(time.Millisecond * 1000)
	url, _ = url.Parse("https://api.twitter.com/1.1/videos/tweet/config/1096941411180707840.json")
	request = &http.Request{
		Method: "GET",
		URL:    url,
		Header: http.Header{"user-agent": []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"}, "accept-encoding": []string{"gzip", "deflate", "br"}, "authorization": []string{"Bearer " + bearer}},
	}
	resp, err = client.Do(request)
	if err != nil {
		logger.Fatalln("Get video parameters ", err)
	}
	logger.Infoln("x-rate-limit-limit:", resp.Header.Get("x-rate-limit-limit"))
	logger.Infoln("x-rate-limit-remaining:", resp.Header.Get("x-rate-limit-remaining"))
	logger.Infoln("x-rate-limit-reset:", resp.Header.Get("x-rate-limit-reset"))
	body, err := gzip.NewReader(resp.Body)
	if err != nil {
		logger.Fatalln(err)
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		logger.Fatalln(err)
	}
	var videoURL Track
	logger.Infof("%s", b)
	if err := json.Unmarshal(b, &videoURL); err != nil {
		logger.Fatalln(err)
	}
	time.Sleep(time.Millisecond * 1000)
	videDescription, err := client.Get(videoURL.Track.PlaybackURL)
	if err != nil {
		logger.Fatalln("Error PlaybackURL:", videoURL.Track.PlaybackURL, err)

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
			convert := exec.Command("ffmpeg", "-i "+"https://video.twimg.com"+t, "-c:v libx264", "-c:a copy result.mp4")
			convert.Run()
			if err != nil {
				logger.Fatalln("ffmpeg:", err)
			}
			break
		}
	}
}
