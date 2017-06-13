# Spatial information for PNG images

A simple command line tool for calculating spatial information for images based on Soble operator.

## Status

Warning: this tool is very alpha.

# Introduction

The command line tool calculates spatial information based on the Sobel operator. The
spatial information mean value (`SImean`) is printed to stdout.

Spatial information is defined as

<pre>
    SIr = sqrt(sX^2 + sY^2)
</pre>

where `sX` and `sY` are output of the Sobel kernel.

`SImean` is defined as

<pre>
    (1 / pixel) * sum(SIr)
</pre>

where `pixel` is the amount of pixels in the image.

# Usage

Use 

<pre>
   si &lt;image>
</pre>

from command line where `&lt;image>` is the filename of the PNG or JPEG image to get `SImean` for.

# Build

Precondition: installed `go` 1.8.3

Clone the repository and type

<pre>
    go build src/si.go 
</pre>

