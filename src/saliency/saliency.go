package saliency

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
    "lab"
)

type SaliencyMap []float64


func gaussianSmooth(srcImg []float64, width int, height int, kernel []float64, ch chan<- []float64) {
    smoothImg := make([]float64, len(srcImg))
    tmpImg := make([]float64, len(srcImg))
    center := len(kernel) / 2
    
    rows := height
    cols := width

    // Blur in the x direction.    
    idx := 0
    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            var kernelsum float64
            var sum float64
            for cc := 	(-center); cc <= center; cc++ {
                if ((c + cc) >= 0) && ((c + cc) < cols) {
                    sum += srcImg[r * cols + (c + cc)] * kernel[center + cc]
                    kernelsum += kernel[center + cc]
                }
            }
            tmpImg[idx] = sum / kernelsum
            idx++
        } 
    }

    // Blur in the y direction.
    idx = 0
    for r := 0; r < rows; r++ {
        for c := 0; c < cols; c++ {
            var kernelsum float64
            var sum float64
            
            for rr := (-center); rr <= center; rr++ {
                if ((r + rr) >= 0) && ((r + rr) < rows) {
                    sum += tmpImg[(r + rr) * cols + c] * kernel[center + rr]
                    kernelsum += kernel[center + rr]
                }   
            }            
            
            smoothImg[idx] = sum / kernelsum;
            idx++;
        } 
    }

    ch <- smoothImg 
}

func createIntegralImage(srcImg []float64, width int, height int, ch chan<- [][]float64) {
    intImg := make([][]float64, height)
    for i := range intImg {
        intImg[i] = make([]float64, width)
    }
    idx := 0
    
    for j := 0; j < height; j++ {
        var sumRow float64
        for k := 0; k < width; k++ {
            sumRow += srcImg[idx]
            idx++
            if 0 == j {
                intImg[j][k] = sumRow
            } else {
                intImg[j][k] = intImg[j - 1][k] + sumRow
            }
        }
    }
    
    ch <- intImg
}

func getIntegralSum(intImg [][]float64, x1, y1, x2, y2 int) float64 {
    var sum float64
    
    if x1 - 1 < 0 && y1 - 1 < 0 {
        sum = intImg[y2][x2]
    } else if x1 - 1 < 0 {
        sum = intImg[y2][x2] - intImg[y1 - 1][x2]
    } else if y1 - 1 < 0 {
        sum = intImg[y2][x2] - intImg[y2][x1 - 1]
    } else {
        sum = intImg[y2][x2] + intImg[y1 - 1][x1 - 1] - intImg[y1 - 1][x2] - intImg[y2][x1 - 1]
    }
    
    return sum
}

func normalize(salMap SaliencyMap) {
    maxValue := 0.0
    minValue := float64(1 << 30)
    
    size := len(salMap)
    
    for i := 0; i < size; i++ {
        if maxValue < salMap[i] {
            maxValue = salMap[i]
        }
        if minValue > salMap[i] {
            minValue = salMap[i]
        }
    }
    
    
    _range := maxValue - minValue
    if _range <= 0 {panic("Range lower 0")}
    
    for i := 0; i < size; i++ {
        salMap[i] = ((255.0 * (salMap[i] - minValue)) / _range)
    }
    
}

func max(a, b int) int {
    if a <= b {
        return b
    }
    return a
}

func min(a, b int) int {
    if a <= b {
        return a
    }
    return b
}

func MaximumSymmetricSurroundSaliency(source lab.LAB) SaliencyMap {
    width := source.Stride()
    height := len(source.L) / width 
    size := width * height
    saliencyMap := make(SaliencyMap, size)
    
    kernel := []float64{1.0, 2.0, 1.0}
    
    chLs,   chAs,   chBs   := make(chan []float64),    make(chan []float64),    make(chan []float64)
    chLint, chAint, chBint := make(chan [][]float64),  make(chan [][]float64),  make(chan [][]float64)
    go gaussianSmooth(source.L, width, height, kernel, chLs)
    go gaussianSmooth(source.A, width, height, kernel, chAs)
    go gaussianSmooth(source.B, width, height, kernel, chBs)

    go createIntegralImage(source.L, width, height, chLint)
    go createIntegralImage(source.A, width, height, chAint)
    go createIntegralImage(source.B, width, height, chBint)
    ls := <- chLs
    as := <- chAs
    bs := <- chBs
    lint := <-chLint
    aint := <-chAint
    bint := <-chBint

    index := 0
    for j := 0; j < height; j++ {
        yoff := min(j, height - j)
        y1 := max(j - yoff, 0)
        y2 := min(j + yoff, height - 1)

        for k := 0; k < width; k++ {
            xoff := min(k, width - k)
            x1 := max(k - xoff, 0)
            x2 := min(k + xoff, width - 1)

            area := (x2 - x1 + 1) * (y2 - y1 + 1);
            
            lval := getIntegralSum(lint, x1, y1, x2, y2) / float64(area)
            aval := getIntegralSum(aint, x1, y1, x2, y2) / float64(area)
            bval := getIntegralSum(bint, x1, y1, x2, y2) / float64(area)
            
            saliencyMap[index] = (lval - ls[index]) * (lval - ls[index]) + (aval - as[index]) * (aval - as[index]) + (bval - bs[index]) * (bval - bs[index]); //square of the euclidean distance
            index++
        }
    }
    normalize(saliencyMap)    
    
    return saliencyMap
}
