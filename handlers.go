/*
The MIT License (MIT)

zhe.chen<chenzhe07@gmail.com>
*/

package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kennygrant/sanitize"
	"io"
	"io/ioutil"
	"log"
	_ "math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Approaching Neutral Zone, all systems normal and functioning.")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(404), 404)
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(_24K); nil != err {
		log.Printf("%s", err.Error())
		http.Error(w, "Error occured copying to output stream", 500)
		return
	}

	token := GenRandStr(SYMBOLS)

	w.Header().Set("Content-Type", "text/plain")

	for _, fheaders := range r.MultipartForm.File {
		for _, fheader := range fheaders {
			filename := sanitize.Path(filepath.Base(fheader.Filename))
			contentType := fheader.Header.Get("Content-Type")

			if contentType == "" {
				contentType = mime.TypeByExtension(filepath.Ext(fheader.Filename))
			}

			var f io.Reader
			var err error

			if f, err = fheader.Open(); err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			var b bytes.Buffer

			n, err := io.CopyN(&b, f, _24K+1)
			if err != nil && err != io.EOF {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			var reader io.Reader

			if n > _24K {
				file, err := ioutil.TempFile(config.Temp, "transfer-")
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				n, err = io.Copy(file, io.MultiReader(&b, f))
				if err != nil {
					os.Remove(file.Name())

					log.Printf("%s", err.Error())
					http.Error(w, err.Error(), 500)
					return
				}

				reader, err = os.Open(file.Name())
			} else {
				reader = bytes.NewReader(b.Bytes())
			}

			contentLength := n

			log.Printf("Uploading %s %s %d %s", token, filename, contentLength, contentType)

			if err = storage.Put(token, filename, reader, contentType, uint64(contentLength)); err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return

			}

			fmt.Fprintf(w, "https://%s/%s/%s\n", ipAddrFromRemoteAddr(r.Host), token, filename)
		}
	}
}

func putHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	filename := sanitize.Path(filepath.Base(vars["filename"]))

	contentLength := r.ContentLength

	var reader io.Reader

	reader = r.Body

	if contentLength == -1 {
		// queue file to disk, because s3 needs content length
		var err error
		var f io.Reader

		f = reader

		var b bytes.Buffer

		n, err := io.CopyN(&b, f, _24K+1)
		if err != nil && err != io.EOF {
			log.Printf("%s", err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		if n > _24K {
			file, err := ioutil.TempFile(config.Temp, "transfer-")
			if err != nil {
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			defer file.Close()

			n, err = io.Copy(file, io.MultiReader(&b, f))
			if err != nil {
				os.Remove(file.Name())
				log.Printf("%s", err.Error())
				http.Error(w, err.Error(), 500)
				return
			}

			reader, err = os.Open(file.Name())
		} else {
			reader = bytes.NewReader(b.Bytes())
		}

		contentLength = n
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(vars["filename"]))
	}

	token := GenRandStr(SYMBOLS)

	//log.Printf("Uploading %s %d %s", token, filename, contentLength, contentType)
	log.Printf("upload -- http://%s/%s/%s size:%d type:%s user:%s pass:%s", r.Host, token, filename, contentLength, contentType)

	var err error

	if err = storage.Put(token, filename, reader, contentType, uint64(contentLength)); err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, errors.New("Could not save file").Error(), 500)
		return
	}

	// w.Statuscode = 200

	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "http://%s/%s/%s\n", r.Host, token, filename)
}

func zipHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	zipfilename := fmt.Sprintf("transfersh-%d.zip", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", zipfilename))
	w.Header().Set("Connection", "close")

	zw := zip.NewWriter(w)

	for _, key := range strings.Split(files, ",") {
		if strings.HasPrefix(key, "/") {
			key = key[1:]
		}

		key = strings.Replace(key, "\\", "/", -1)

		token := strings.Split(key, "/")[0]
		filename := sanitize.Path(strings.Split(key, "/")[1])

		reader, _, _, err := storage.Get(token, filename)

		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &zip.FileHeader{
			Name:         strings.Split(key, "/")[1],
			Method:       zip.Store,
			ModifiedTime: uint16(time.Now().UnixNano()),
			ModifiedDate: uint16(time.Now().UnixNano()),
		}

		fw, err := zw.CreateHeader(header)

		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(fw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}

	if err := zw.Close(); err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Internal server error.", 500)
		return
	}
}

func tarGzHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar.gz", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	os := gzip.NewWriter(w)
	defer os.Close()

	zw := tar.NewWriter(os)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		if strings.HasPrefix(key, "/") {
			key = key[1:]
		}

		key = strings.Replace(key, "\\", "/", -1)

		token := strings.Split(key, "/")[0]
		filename := sanitize.Path(strings.Split(key, "/")[1])

		reader, _, contentLength, err := storage.Get(token, filename)
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &tar.Header{
			Name: strings.Split(key, "/")[1],
			Size: int64(contentLength),
		}

		err = zw.WriteHeader(header)
		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(zw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}
}

func tarHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	files := vars["files"]

	tarfilename := fmt.Sprintf("transfersh-%d.tar", uint16(time.Now().UnixNano()))

	w.Header().Set("Content-Type", "application/x-tar")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", tarfilename))
	w.Header().Set("Connection", "close")

	zw := tar.NewWriter(w)
	defer zw.Close()

	for _, key := range strings.Split(files, ",") {
		token := strings.Split(key, "/")[0]
		filename := strings.Split(key, "/")[1]

		reader, _, contentLength, err := storage.Get(token, filename)
		if err != nil {
			if err.Error() == "The specified key does not exist." {
				http.Error(w, "File not found", 404)
				return
			} else {
				log.Printf("%s", err.Error())
				http.Error(w, "Could not retrieve file.", 500)
				return
			}
		}

		defer reader.Close()

		header := &tar.Header{
			Name: strings.Split(key, "/")[1],
			Size: int64(contentLength),
		}

		err = zw.WriteHeader(header)
		if err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}

		if _, err = io.Copy(zw, reader); err != nil {
			log.Printf("%s", err.Error())
			http.Error(w, "Internal server error.", 500)
			return
		}
	}
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]

	reader, contentType, contentLength, err := storage.Get(token, filename)
	if err != nil {
		if err.Error() == "The specified key does not exist." {
			http.Error(w, "File not found", 404)
			return
		} else {
			log.Printf("%s", err.Error())
			http.Error(w, "Could not retrieve file.", 500)
			return
		}
	}

	defer reader.Close()

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatUint(contentLength, 10))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Connection", "close")

	if _, err = io.Copy(w, reader); err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "Error occured copying to output stream", 500)
		return
	}
}

func delHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	token := vars["token"]
	filename := vars["filename"]

	contentType, err := storage.Del(token, filename)
	if err != nil {
		log.Printf("%s", err.Error())
		http.Error(w, "delete file error.", 500)
		return
	}
	log.Printf("delete -- http://%s/%s/%s user:%s pass:%s", ipAddrFromRemoteAddr(r.Host), token, filename)

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("delete attachment; file=\"%s/%s\"", token, filename))
	w.Header().Set("Connection", "close")
	fmt.Fprintf(w, "delete %s/%s ok\n", token, filename)
}

func RedirectHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*	if r.URL.Path == "/health.html" {
			} else if ipAddrFromRemoteAddr(r.Host) == "127.0.0.1" {
			} else if strings.HasSuffix(ipAddrFromRemoteAddr(r.Host), ".elasticbeanstalk.com") {
			} else if ipAddrFromRemoteAddr(r.Host) == "jxm5d6emw5rknovg.onion" {
			} else if ipAddrFromRemoteAddr(r.Host) == "transfer.sh" {
				if r.Header.Get("X-Forwarded-Proto") != "https" && r.Method == "GET" {
					//http.Redirect(w, r, "https://transfer.sh"+r.RequestURI, 301)
					return
				}
			} else if ipAddrFromRemoteAddr(r.Host) != "transfer.sh" {
				//http.Redirect(w, r, "https://transfer.sh"+r.RequestURI, 301)
				return
			}
		*/

		h.ServeHTTP(w, r)
	}
}

// Create a log handler for every request it receives.
func LoveHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-made-with", "<3 by DutchCoders")
		w.Header().Set("x-served-by", "Proudly served by DutchCoders")
		w.Header().Set("Server", "Transfer.sh HTTP Server 1.0")
		h.ServeHTTP(w, r)
	}
}
