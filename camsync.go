/*
  Copyright 2019 Google LLC

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

package camsync

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/scottlaird/virb"
)

func NewCamsync(host string, dir string, mirror bool, pollSeconds int, deletePercent int) *Camsync {
	c := Camsync{
		host:          host,
		dir:           dir,
		mirror:        mirror,
		pollSeconds:   pollSeconds,
		deletePercent: deletePercent,
	}

	return &c
}

type Camsync struct {
	host   string
	dir    string // Local directory for writing downloaded files into
	mirror bool   // If true, copy all files that Garmin references
	// in mediaList and match the directory
	// structure.  Otherwise only copy the file
	// referenced by 'Url' and write it out as
	// 'Name'.
	pollSeconds int // number of seconds to wait between polling
	// attempts.  If 0, then only poll once and then exit.
	deletePercent int

	// TODO(laird): add deletion logic.  Delete immediately?  Delete oldest after x% full?
}

// Map a source name to a destination name, using the mirror and
// output dir settings in *Camsync.
func (c *Camsync) filemap(name, url string) string {
	// If mirror is set, then we extract the file name from the
	// URL, keeping everything starting with /DCIM/.  This should
	// be enough for Garmin's tools (like the Virb 360's editor)
	// to import the video without hassles.
	//
	if c.mirror {
		i := strings.Index(url, "/DCIM/")
		if i < 0 {
			i = strings.Index(url, "/GMetrix/")
		}
		base := url
		if i > 0 {
			base = url[i:]
		}
		return filepath.Clean(filepath.Join(c.dir, base))
	}

	// Otherwise, just use the name provided.
	return filepath.Clean(filepath.Join(c.dir, name))
}

// Big files.  Don't read them entirely into RAM.
func mirrorfile(url, filename string) error {
	if url == "" || filename == "" {
		return nil
	}

	glog.Infof("Fetching %s into %s", url, filename)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	glog.Infof("Server says ContentLength of %d", resp.ContentLength)

	// Make sure output directory exists
	dirname := filepath.Dir(filename)
	err = os.MkdirAll(dirname, 0755)
	if err != nil && !os.IsNotExist(err) {
		glog.Errorf("Error creating output directory: %v", err)
		return err
	}

	// Check to see if the output file exists and see how big it is.
	stat, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		glog.Errorf("Error statting output file: %v", err)
		return err
	}
	// If the sizes don't match, then download.
	download := false
	if err != nil {
		download = true
		glog.Info("File missing locally; downloading")
	}
	if err == nil && stat.Size() != resp.ContentLength {
		download = true
		glog.Info("File size mismatch; re-downloading")
	}

	if download {
		start := time.Now()
		glog.Infof("Downloading %s into %s", url, filename)

		o, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer o.Close()

		size, err := io.Copy(o, resp.Body)
		if err != nil {
			glog.Errorf("Copy failed: %v", err)
			return err
		}
		done := time.Now()
		elapsed := done.Sub(start).Seconds()
		glog.Infof("Copied %d bytes in %.1f seconds (%.1f kB/sec)", size, elapsed, float64(size)/elapsed/1000)
	} else {
		glog.Info("Skipping download; file exists.")
	}
	return nil
}

func (c *Camsync) Sync() error {
	glog.Infof("Fetching media list from %s", c.host)
	r, err := virb.MediaList(c.host, "")
	if err != nil {
		return err
	}

	glog.Infof("MediaList returned %d items", len(r.Media))

	for _, m := range r.Media {
		glog.Infof("Found at %s: %#v", m.Name, m)

		// Produce a map of URL->file mappings for files that
		// we might need to fetch.
		f := make(map[string]string)
		f[m.Url] = c.filemap(m.Name, m.Url)
		if c.mirror {
			f[m.FitURL] = c.filemap("", m.FitURL)
			f[m.LowResVideoPath] = c.filemap("", m.LowResVideoPath)
			f[m.ThumbURL] = c.filemap("", m.ThumbURL)
		}

		for k, v := range f {
			err = mirrorfile(k, v)
			if err != nil {
				glog.Errorf("File mirroring failed on %s: %v", k, err)
			}
		}
	}
	return nil
}

func (c *Camsync) Wait() {
	time.Sleep(time.Duration(c.pollSeconds) * time.Second)
}
