# inpxer

OPDS 1.1 and web server for `.inpx` libraries with full-text search.

## Usage

### Standalone

Download the latest release.
Download [`inpxer-example.toml`](./inpxer-example.toml), rename to `inpxer.toml`, put near executable and edit to your liking.

Import data:
```shell
./inpxer import ./file.inpx
```

*Note: existing index will be deleted.*

Start server:
```shell
./inpxer serve
```

Web interface will be available on [http://localhost:8080/](http://localhost:8080/) and
OPDS will be on [http://localhost:8080/opds](http://localhost:8080/opds) by default.

### Docker

Download [`inpxer-example.toml`](./inpxer-example.toml), rename to `inpxer.toml` and edit to your liking.

inpxer expects config file to be at `/data/inpxer.toml`.

Import data:
```shell
docker run --rm -it -v ${PWD}:/import -v <path to data storage>:/data shemanaev/inpxer inpxer import /import/file.inpx
```

*Note: existing index will be deleted.*

Start server:
```shell
docker run -it -p 8080:8080 -v <path to data storage>:/data shemanaev/inpxer
```
