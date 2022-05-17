# httpzipdir

An http server that serves directories like the
[autoindex](https://nginx.org/en/docs/http/ngx_http_autoindex_module.html) module in nginx and
[dirlisting](https://redmine.lighttpd.net/projects/1/wiki/Docs_ModDirlisting) in lighttpd.

However, for each subdirectory, it adds a "dynamic" `subdirectory.zip` file in the listing.
When such a file is requested, `httpzipdir` will stream the content of the subdirectory as a zipfile to
the client.

The index template was lifted mostly verbatim from the output of lighttpd's dirlisting module.

If a requested directory contains an `index.html` file, the content of the file is sent to the client
instead of the automatic index.

Both the `autoindex` and `zipdir` features can be turned off on a per "allowed" directory basis.
See the usage message for examples.

# Demo

<https://raby.sh/demos/httpzipdir>

# Usage

```
Usage of ./build/httpzipdir:
      --allow stringToString   Add an allowed url prefix->docroot:options mapping. e.g /media/patate/url=/path/to/media/patate, /url/path=/path/on/disk:noautoindex. Valid options: noautoindex,nozipdir (default [])
      --listen string          Listen address (default "127.0.0.1")
      --port int               Listen port (default 10666)
      --version                Show version and exit
```
