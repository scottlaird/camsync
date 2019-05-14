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

package main

import (
	"flag"
	"os"

	"github.com/golang/glog"
	"github.com/scottlaird/camsync"
)

var hostname = flag.String("camera", "10.1.0.172", "Hostname or IP address of Garmin Virb")
var output_directory = flag.String("output_directory", "../DATA", "Output directory")
var mirror = flag.Bool("mirror", true, "If true, mirror Garmin's directory structure and all files.  If false, then only copy media.")
var poll = flag.Bool("poll", false, "If true, keep running and attempting to resync periodically.")
var pollInterval = flag.Int("poll_interval", 900, "Number of seconds to wait between polling for new camera content.")
var deletePercent = flag.Int("delete_percent", 0, "Delete files if the camera is over X percent full.  0 disables deletes.")

func main() {
	flag.Parse()

	glog.Info("Starting")
	pollSeconds := 0
	if *poll {
		pollSeconds = *pollInterval
	}
	c := camsync.NewCamsync(*hostname, *output_directory, *mirror, pollSeconds, *deletePercent)

	for {
		ret := 0
		err := c.Sync()
		if err != nil {
			glog.Error(err)
			ret = 1
		}
		if !*poll {
			os.Exit(ret)
		}
		c.Wait()
	}
}
