# fedora-coreos-stream-generator
Generates stream metadata for Fedora CoreOS using release metadata and overrides.

# Running
Make sure you have the latest Go installed and then download the package.
```
$ go get -u github.com/coreos/fedora-coreos-stream-generator
```
Add the binary in your PATH variable. If you haven't done anything fancy, adding the standard binary path where Go keeps binaries from packages will be enough.
```
$ export PATH=$PATH:~/go/bin/
$ fedora-coreos-stream-generator -releases=<release index location for the stream> -output-file=stream.json
```
A partial stream override can be specified with `-override=</path/to/override.json>`. `override.json` needs to be available locally on the system.

# Development
Fork and clone the repo locally and run it:
```
$ make
$ ./fedora-coreos-stream-generator -releases=<release index location for the stream>
```

Make sure to run `make update` to keep dependencies up-to-date.

Make changes and send a PR!
