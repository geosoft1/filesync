// filesync project [client]
// Copyright (C) 2017  geosoft1  geosoft1@gmail.com
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Config struct {
	Ip       string `json:"ip"`
	Port     string `json:"port"`
	Path     string `json:"path"`
	SyncTime string `json:"synctime"`
	Addr     string
}

var c Config
var pwd string
var err error

type SyncMask struct {
	Filename string
	DateTime time.Time
}

var LocalSyncMask = make([]SyncMask, 0, 1)
var RemoteSyncMask = make([]SyncMask, 0, 1)
var SyncFileList = make([]string, 0, 1)

var r *http.Response
var out *os.File

func HandleFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		relPath, _ := filepath.Rel(c.Path, path)
		//very important for crossplatform operations
		relPath = filepath.ToSlash(relPath)
		LocalSyncMask = append(LocalSyncMask, SyncMask{relPath, info.ModTime()})
	}
	return nil
}

func GetLocalSyncMask(Path string) {
	LocalSyncMask = nil
	if err := filepath.Walk(Path, HandleFile); err != nil {
		log.Println(err)
	}
}

func GetRemoteSyncMask(addr string) {
	r, err = http.Get("http://" + addr)
	if err != nil {
		log.Println(err)
	} else {
		json.NewDecoder(r.Body).Decode(&RemoteSyncMask)
		r.Body.Close()
	}
}

func Find(a *[]SyncMask, s string) *SyncMask {
	for _, r := range *a {
		if r.Filename == s {
			return &r
		}
	}
	return nil
}

func CompareSyncMasks(local *[]SyncMask, remote *[]SyncMask, filelist *[]string) {
	*filelist = nil
	for _, r := range *remote {
		if tmp := Find(local, r.Filename); tmp == nil || r.DateTime.Sub(tmp.DateTime) > 0 {
			//if file not exist in local list or newest add in the sync list
			*filelist = append(*filelist, r.Filename)
			//print filename to synchronize
			log.Println("-", r.Filename)
		}
	}
}

func DownloadFiles(addr string, path string, filelist *[]string) {
	basename := filepath.ToSlash(path) + string(filepath.Separator)
	os.MkdirAll(basename, os.ModePerm)
	for _, filename := range *filelist {
		t := &url.URL{Path: filename}
		r, _ = http.Get("http://" + addr + "/files/" + t.String())
		dirname := basename + filepath.ToSlash(filepath.Dir(filename))
		os.MkdirAll(dirname, os.ModePerm)
		out, _ = os.Create(basename + filename)
		io.Copy(out, r.Body)
		r.Body.Close()
		out.Close()
		// print syncronized filename
		log.Println("*", filename)
	}
}

func main() {
	log.Print("init logger")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	pwd, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		os.Exit(1)
	}

	log.Print("load configuration")
	f, _ := os.Open(filepath.ToSlash(pwd + "/conf.json"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println(c)

	c.Addr = c.Ip + ":" + c.Port
	t, _ := strconv.Atoi(c.SyncTime)

	for range time.NewTicker(time.Duration(t) * time.Second).C {
		GetLocalSyncMask(c.Path)
		GetRemoteSyncMask(c.Addr)
		CompareSyncMasks(&LocalSyncMask, &RemoteSyncMask, &SyncFileList)
		DownloadFiles(c.Addr, c.Path, &SyncFileList)
	}
}
