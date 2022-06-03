# Makefile

This is the `Makefile` written in markdown for `makedown`.

This `Makefile.md` file can be run by `makedown`.

## Directives for `makedown:`

Note that this `Makefile.md` assumes that you have `gmake` 3.82+ to use `.ONESHELL`.

```
#!/usr/local/bin/gmake -f

.ONESHELL:
```

Make sure you have a required version of `make`,
especially on MacOS. Then, type:

`export MAKEDOWN_MAKE_COMMAND=gmake`

## Howo to generate Go sources by `godown:`

This generates `*.go` from `*.go.md` by [`godown`](https://github.com/hirokistring/godown).

```
godown
```

## Check differences by `diff:`

```
diff main.go main.go.expected || true
diff makedown.go makedown.go.expected || true
```

## How to `generate:`

This only generates `Makefile`. No targets will be executed.

```
./makedown --out Makefile
```

## How to just `build:`

```
go build
```

## Build with godown `build-with-godown:`

```
godown build
```

The generated `makedown` executable file is not ready to be released. It has to be notarized by [gon](https://github.com/mitchellh/gon) for MacOS.

## How to `build-and-notarize:`

This makes binaries for each platform. Then, it signs and notarizes the binary for MacOS.

```
goreleaser build --snapshot --rm-dist
```

### How to `check-notarized:`

```
@echo Check the binary is signed
codesign --display -vvv dist/macos_darwin_amd64_v1/makedown

@echo
@echo Check the binary is notarized
spctl --assess --type install -vvv dist/macos_darwin_amd64_v1/makedown
```

Note that `makedown` is just a binary, not an .app.

## How to try `build-and-release` with `build-and-release-snapshot:`

Type the command bellow.

```sh
goreleaser release --snapshot --rm-dist
```

## How to`build-and-release:`

Type the command bellow.

```sh
goreleaser release --rm-dist
```

Also you can run this `release` target by `makedown` itself, like `$ makedown release`.

## Build and Release Tools

`makedown` uses [goreleaser](https://github.com/goreleaser) to build and release the binary.

`makedown` uses [gon](https://github.com/mitchellh/gon) to notarize the binary for Mac.

## Tests

Make sure that you have a make command 3.82+ before running the tests.

### `case1:`

```
cd tests/case1
if [ -e Makefile ]; then
  rm Makefile
fi
../../makedown --out Makefile
diff Makefile Makefile.expected > Makefile.diff
cat Makefile.diff
```

### `case2:`

```
cd tests/case2
if [ -e Makefile ]; then
  rm Makefile
fi
../../makedown -f EXAMPLE.md
diff Makefile Makefile.expected > Makefile.diff
cat Makefile.diff
```

## Tests for README.md

The example targets in `README.md` are copied here, for users cloned this repository first time.
That is because `makedown` gives precedence to `Makefile.md` over `README.md`.

### `sayhello:`

```sh
@echo "Hello, $(WHO)!"
```

### `saygoodbye:`

```sh
@echo "Bye for now," `date +%Y/%m/%d`
```

### `variables:`

```makefile
WHO = makedown
```

### `install:`

```
go install github.com/hirokistring/makedown
```
