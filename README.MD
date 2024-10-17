# GoPizer

GoPizer is a Golang image optimizer service that can optimize images by resizing, compressing, and converting them to the best format to serve based on which formats the browser supports.

GoPizer uses [libvips](https://www.libvips.org/) for image processing that is a very fast and efficient image processing library with many image format support through [GoVips](https://pkg.go.dev/github.com/davidbyttow/govips/v2/vips) bindings.

GoPizer can store the optimized images in the local file system, in memory, in a [Redis](https://redis.io/) or [Memcache](https://memcached.org/) cache.

## Quickstart

To install GoPizer, you can use the following command: