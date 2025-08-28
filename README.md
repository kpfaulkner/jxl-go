# jxl-go

[![Go Reference](https://pkg.go.dev/badge/github.com/kpfaulkner/jxl-go.svg)](https://pkg.go.dev/github.com/kpfaulkner/jxl-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kpfaulkner/jxl-go)](https://goreportcard.com/report/github.com/kpfaulkner/jxl-go)
[![Go Coverage](https://github.com/kpfaulkner/jxl-go/wiki/coverage.svg)](https://raw.githack.com/wiki/kpfaulkner/jxl-go/coverage.html)
[![Go](https://github.com/kpfaulkner/jxl-go/actions/workflows/ci.yaml/badge.svg)](https://github.com/kpfaulkner/jxl-go/actions/workflows/ci.yaml)


## Notes:

JXL-Go is a Go library for reading of JPEG XL (JXL) images.
Currently the focus is on reading lossless images but will expand to lossy images in the future.

This started off based off the JXL specs, JXL reference implementation (https://github.com/libjxl/libjxl) , JXLatte ( https://github.com/Traneptora/jxlatte ) 
and jxl-oxide ( https://github.com/tirr-c/jxl-oxide ). These days it's probably better described as 90% a Go port of JXLatte.

## Example usage:

```go

  // read JXL file 
  file := `../testdata/lossless.jxl`
  f, _ := os.ReadFile(file)
  r := bytes.NewReader(f)
  jxl := core.NewJXLDecoder(r,nil)
  jxlImage, _ := jxl.Decode()

  // convert to regular Go image.Image
  img, _ := jxlImage.ToImage()
```

## Example images:

A number of good test JXL images are located in the testdata directory which exercise different parts of the 
decoding process. In particular lossless.jxl, church.jxl and lenna.jxl are good examples.

## TODO:

- [ ] Performance improvements
- [ ] Refactor into appropriate modules/packages
- [ ] Add tests
- [ ] Remove THE MANY unnecessary type casting (have lots for int conversions)
- [ ] Remove all panics (currently using them to indicate sections not implemented)

## Performance

The focus is currently on being able to read the JXL image and generating a Go Image struct for applications to use.
Once that is done will focus on performance improvements. There are many areas to look at:

- Memory allocation
- Parallelism
- SIMD

## Notes

Good to generate PFM files then use a site like https://imagetostl.com/convert/file/pfm/to/png#convert to 
convert to PNG for easier viewing.

To generate coverage report:
```
go test -v -coverprofile cover.out ./...
go tool cover -html cover.out -o cover.html
```