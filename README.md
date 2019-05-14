# camsync

This is designed to download photos and video from Garmin Virb action
cameras over WiFi using the [Virb
API](https://developer.garmin.com/downloads/virb/Camera_Network_Services_API_v0.5.pdf).

It has been tested with a Virb Ultra 30 and a Virb 360.  It'll
probably also work with the Virb X, Virb XE, and Garmin's Dashcams,
but these haven't been tested.

It can be run in two modes: `--mirror` and `--nomirror`.  If `--mirror` is
used, then it will attempt to reproduce the file structure from the
Garmin camera and copy all of the metadata files referenced in the
camera's media list, including `.FIT` files.  This *should* be enough
for it to work with VirbEdit, but it hasn't been tested yet.  When run
in `--nomirror` mode, the program will only download the primary image
files (`.JPG` and `.MP4`) and will ignore `.FIT`, `.THM`, `.GLV`, and other
files.  In `--nomirror` mode, all downloaded files will be written to a
single directory, and no attempt will be made to reproduce Garmin's
directory structure.

In short, if you just want the videos, use `--nomirror`.  If you want to
feed the videos to VirbEdit or similar tools, then use `--mirror`.

`--output_directory` tells camsync which local directory to write to.

`--camera` is the camera's hostname or IP address.

Use `--logtostderr` to get additional debugging data.

### Disclaimer

This is not an officially supported Google product.
