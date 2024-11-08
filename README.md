# jxl-go

## Notes:

JXL-Go is a Go library for reading of JPEG XL (JXL) images.
Currently the focus is on reading lossless images but will expand to lossy images in the future.

This started off based off the JXL specs, JXL reference implementation (https://github.com/libjxl/libjxl) , JXLatte ( https://github.com/Traneptora/jxlatte ) 
and jxl-oxide ( https://github.com/tirr-c/jxl-oxide ). These days it's probably better described as 90% a Go port of JXLatte.


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