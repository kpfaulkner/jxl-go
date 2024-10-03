# jxl-go

## Notes:

JXL-Go is a Go library for reading of JPEG XL (JXL) images.
Currently the focus is on reading lossless images but will expand to lossy images in the future.

This is based off the JXL specs, JXL reference implementation (https://github.com/libjxl/libjxl) and the JXLatte ( https://github.com/Traneptora/jxlatte ) project

Currently JXLatte takes about 397ms to decode test image where as JXL-Go is currently taking 650ms.


## TODO:

- [ ] Performance improvements
- [ ] Refactor into appropriate modules/packages
- [ ] Add tests
- [ ] Remove unnecessary type casting (have lots for int conversions)
- [ ] Remove all panics (currently using them to indicate sections not implemented)
- 
## Performance

The focus is currently on being able to read the JXL image and generating a Go Image struct for applications to use.
Once that is done will focus on performance improvements. There are many areas to look at:

- Memory allocation
- Convert 2D and 3D slices into 1D slice but with helper functions to access 2D and 3D slices
- IO (currently lots of single byte reads/writes)
- Parallelism
- SIMD

## Notes

