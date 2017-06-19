package main

/*
 * Computes the saliency mean factor based on Radhakrishna Achanta and and S. Suesstrunk
 * salient detection algorithm published in "Saliency Detection using Maximum Symmetric Surround",
 * Proceedings of IEEE International Conference on Image Processing (ICIP), 2010.
 *
 * This code is a simple scripted go implementation of the C++ code, published at
 * http://ivrl.epfl.ch/page-75168-en.html
 *
 * Original code (c) 2010 Radhakrishna Achanta [EPFL]. All rights reserved.
 * Go code (c) 2017 Sascha Kohlmann
 */

import (
    "flag"
    "fmt"
    "image"
    _ "image/jpeg" // register the PNG format with the image package
    "image/png" // register the PNG format with the image package
    "os"
    "strings"
    "time"
    "lab"
    "saliency"
)

func loadImage(name string) image.Image {
    infile, err := os.Open(name)
    check(err)
    defer infile.Close()
    
    src, _, err := image.Decode(infile)

    check(err)
    return src
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

var verbose *bool

func verbosePrintExecDuration(t time.Time, prefix string) {
    if *verbose {
        fmt.Fprintf(os.Stderr, "%s: %d\n", prefix, time.Since(t).Nanoseconds())
    }
}

func main() {
    if len(os.Args) == 1 {
        usage(os.Args[0])
        return
    }

    saliencyImgName := flag.String("o", "", "Name of the saliency image")
    help := flag.Bool("h", false, "Prints this help")
    verbose = flag.Bool("v", false, "Prints verbose informtion")
    flag.Parse()
    srcImgName := os.Args[len(os.Args) - 1]

    if *help {
        usage(os.Args[0])
        return
    }

    startLoad := time.Now()
    src := loadImage(srcImgName)
    verbosePrintExecDuration(startLoad, "load image")
    
    startLab := time.Now()
    lab := lab.ImageToLab(src)
    verbosePrintExecDuration(startLab, "LAB transformation")

    bounds := src.Bounds()
    startSal := time.Now()
    salMap := saliency.MaximumSymmetricSurroundSaliency(*lab)
    verbosePrintExecDuration(startSal, "saliency calculation")

    gray := image.NewGray(bounds)
    size := len(salMap)
    var Ssum float64
    for i := 0; i < size; i++ {
        Ssum += salMap[i]
        gray.Pix[i] = uint8(salMap[i] * 1.1)
    }
    
    Smean := (1.0 / float64((float64(bounds.Max.X * bounds.Max.Y))) * Ssum)
    fmt.Printf("%f", Smean)
    
    if strings.Compare(*saliencyImgName, "") != 0 {
        saveFile, err := os.Create(*saliencyImgName)
        check(err)
        defer saveFile.Close()
        err = png.Encode(saveFile, gray)
        check(err)
    }
}

func header() {
    fmt.Fprintf(os.Stderr, "Saliency Detection using Maximum Symmetric Surround of images.\n")
    fmt.Fprintf(os.Stderr, "Copyright (c) 2017 Sascha Kohlmann.\n")
}

func usage(prgName string) {
    fmt.Fprintf(os.Stderr, "usage: %s [options] <image>\n\n", prgName)
    header()
    fmt.Fprintf(os.Stderr, "\nOptions:\n")
    fmt.Fprintf(os.Stderr, "  -h       : prints this help\n")
    fmt.Fprintf(os.Stderr, "  -o name  : stores a saliency control image with <name> - optional\n")
    fmt.Fprintf(os.Stderr, "  -v       : verbose\n")
}
