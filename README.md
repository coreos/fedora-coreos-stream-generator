# fedora-coreos-stream-generator
Generates stream metadata for Fedora CoreOS using release metadata and overrides

# Running
Make sure you have latest go installed and then download the pacakge
```
$ go get -u github.com/sinnykumari/fedora-coreos-stream-generator
```
Add the binary in your PATH variable. If you haven't done anything fancy, adding standard binary path where Go keeps binaries from package will be enough
```
$ export PATH=$PATH:~/go/bin/
$ fedora-coreos-stream-generator -release-index=<release index location for the stream>
```
Partial stream override can be given using option -override=<override json file location for the stream> . override.json need to be available locally to the system. https://github.com/coreos/fedora-coreos-streams repo will contains override for FCOS streams

**Note:** We don't yet generate release index (releases.json) for FCOS stream.  https://sinnykumari.fedorapeople.org/fcos/release_index.json See https://github.com/coreos/fedora-coreos-tracker/issues/98 for sample example

# Development
Fork and clone the repo locally and run it
```
$ go build .
$ ./fedora-coreos-stream-generator -releases-index=<release index location for the stream>

```

Make changes and send PR

