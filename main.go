package main

import (
	"log"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"github.com/valyala/tcplisten"
)

var reurl = regexp.MustCompile(`\/(\d*)$`)

func init() {
	logger.SetFormatter(&logrus.JSONFormatter{})
}

func main() {
	e := fasthttprouter.New()
	e.GET("/*video", func(c *fasthttp.RequestCtx) {
		url := reurl.FindSubmatch([]byte(c.UserValue("video").(string)))
		if url != nil {
			filename, err := converter(string(url[1]))
			if err != nil {
				c.Response.AppendBody([]byte(err.Error()))
				c.SetStatusCode(400)
				return
			}
			c.SetContentType("video/x-matroska")
			if err := c.Response.SendFile("./videos/" + filename + ".mkv"); err != nil {
				c.Response.AppendBody([]byte("Error"))
				c.SetStatusCode(400)
				return
			}
			c.SetStatusCode(200)
		}
	})

	cfg := tcplisten.Config{
		ReusePort:   true,
		FastOpen:    true,
		DeferAccept: true,
		Backlog:     1024,
	}
	ln, err := cfg.NewListener("tcp4", ":8080")
	if err != nil {
		log.Fatalf("error in reuseport listener: %s\n", err)
	}
	serv := fasthttp.Server{Handler: e.Handler, ReduceMemoryUsage: false, Name: "highload", Concurrency: 2 * 1024, DisableHeaderNamesNormalizing: true}
	if err := serv.Serve(ln); err != nil {
		log.Fatalf("error in fasthttp Server: %s", err)
	}

}
