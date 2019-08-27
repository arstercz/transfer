/*
The MIT License (MIT)

Copyright (c) 2014 DutchCoders [https://github.com/dutchcoders/]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package main

import (
	// _ "transfer.sh/app/handlers"
	// _ "transfer.sh/app/utils"
	"flag"
	"fmt"
	"github.com/PuerkitoBio/ghost/handlers"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"time"
)

const SERVER_INFO = "transfer"

var IsAuth bool = false

// parse request with maximum memory of _24Kilobits
const _24K = (1 << 20) * 24

var config struct {
	Temp string
	Conf string
}

var storage Storage

func init() {
	config.Temp = os.TempDir()
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	port := flag.String("port", "8080", "port number, default: 8080")
	temp := flag.String("temp", config.Temp, "")
	basedir := flag.String("basedir", "", "")
	logpath := flag.String("log", "", "")
	isauth := flag.Bool("auth", false, "whether enable http basic auth or not")
	conf := flag.String("conf", "", "config file to verify nomarl user")
	flag.Parse()

	IsAuth = *isauth
	r := mux.NewRouter()
	r.HandleFunc("/{token}/{filename}", auth(getHandler, basicAuth)).Methods("GET")
	r.HandleFunc("/get/{token}/{filename}", auth(getHandler, basicAuth)).Methods("GET")
	r.HandleFunc("/put/{filename}", auth(putHandler, basicAuth)).Methods("PUT")
	r.HandleFunc("/upload/{filename}", auth(putHandler, basicAuth)).Methods("PUT")
	r.HandleFunc("/{filename}", auth(putHandler, basicAuth)).Methods("PUT")
	r.HandleFunc("/health.html", healthHandler).Methods("GET")
	r.HandleFunc("/", auth(postHandler, basicAuth)).Methods("POST")
	// r.HandleFunc("/{page}", viewHandler).Methods("GET")
	r.HandleFunc("/{token}/{filename}", auth(delHandler, basicAuth)).Methods("DELETE")

	r.NotFoundHandler = http.HandlerFunc(notFoundHandler)

	config.Temp = *temp
	config.Conf = *conf

	if IsAuth && len(config.Conf) <= 0 {
		fmt.Println("error: You must secipfy the conf file to add the username and password!")
		os.Exit(1)
	}
	if *logpath != "" {
		f, err := os.OpenFile(*logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		}

		defer f.Close()

		log.SetOutput(f)
	}

	var err error

	if *basedir == "" {
		log.Panic("basedir not set")
	}
	storage, err = NewLocalStorage(*basedir)

	if err != nil {
		log.Panic("Error while creating storage.", err)
	}

	mime.AddExtensionType(".md", "text/x-markdown")

	log.Printf("Transfer.sh server started. :\nlistening on port: %v\nusing temp folder: %s\n", *port, config.Temp)
	log.Printf("---------------------------")

	s := &http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: handlers.PanicHandler(LoveHandler(RedirectHandler(handlers.LogHandler(r, handlers.NewLogOptions(log.Printf, "_default_")))), nil),
	}

	log.Panic(s.ListenAndServe())
	log.Printf("Server stopped.")
}
