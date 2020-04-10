package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func runCommand() error {
	if isInputFromPipe() {
		print("data is from pipe")
		return aHisto(os.Stdin, os.Stdout)
	}
	//} else {
	file, e := getFile()
	if e != nil {
		return e
	}
	defer file.Close()
	return aHisto(file, os.Stdout)
	//}
}

func isInputFromPipe() bool {
	fi, _ := os.Stdin.Stat()
	return fi.Mode()&os.ModeCharDevice == 0
}

func getFile() (*os.File, error) {
	if flags.filepath == "" {
		return nil, errors.New("please input a file")
	}
	if !fileExists(flags.filepath) {
		return nil, errors.New("the file provided does not exist")
	}
	file, e := os.Open(flags.filepath)
	if e != nil {
		return nil, errors.Wrapf(e,
			"unable to read the file %s", flags.filepath)
	}
	return file, nil
}

func aHisto(r io.Reader, w io.Writer) error {
	// old form: scanner := bufio.NewScanner(bufio.NewReader(r))
	scanner := bufio.NewReader(r)
	var depthLine []string
	var newval int
	var lastval int
	lastval = -1
	var chrSize uint64
	var histsize = 512
	var histogram [512]uint32 // depth histogram for whole chromosome 1 - 511 no 0 as this will be to big
	var maxhisto uint32       // region max of histo from 1 to 510
	var maxzero uint32
	var maxzerogram uint32
	var maxzerocon uint32
	const zerosize = 500
	var zerogram [zerosize]uint32
	maxhisto = 0

	var maxregionhisto uint32
	//	var regionpercenthistogram[512] uint // region % histogram
	var regionName string
	var lastRegion string
	lastRegion = "_last_region"

	for {
		// text expected "chr1  1   1" region, position , readdepth
		// (if -a specified in samtools depth the position will be 0- length of region) otherwise 0 depths are omitted from output)

		// depthLine = strings.Fields(scanner.Text())
		line, err := (scanner.ReadString(10)) // read up to eol
		if err == io.EOF {
			fmt.Println("finshed reading input")
			break
		}
		// get file line current value to add to histogram field 0 or 3
		depthLine = strings.Fields(line)
		switch {
		case len(depthLine) == 1:
			newval, err = strconv.Atoi(depthLine[0])
			regionName = "file"
			lastRegion = regionName
		case len(depthLine) == 3:
			newval, err = strconv.Atoi(depthLine[2])
			regionName = depthLine[0]
			if lastRegion == "_last_region" {
				lastRegion = regionName
			}
		case len(depthLine) == 0:
			continue // empty line, go read next line ?
		default:
			newval = -1
		}
		if err != nil {
			fmt.Println(" data non numeric/integer, exiting")
			os.Exit(-1)
		}
		// end get data from file line
		// now test if we have new region
		if regionName != lastRegion {
			for i := 1; i < (histsize - 1); i++ {
				if histogram[i] > maxhisto {
					maxhisto = histogram[i]
				}
			}
			printRegion(lastRegion, histogram[:], chrSize, maxhisto, maxregionhisto, histsize)
			maxhisto = 0
			maxregionhisto = 0
			
			for i := range histogram {
				histogram[i] = 0
			}
			fmt.Println("><><>< zerogram: ")
			printRegion(lastRegion, zerogram[:], chrSize, maxzerogram, maxzerocon, zerosize)
			for i := range zerogram {
				zerogram[i] = 0
			}
			maxzero = 0
			maxzerocon = 0
			maxzerogram = 0
			lastval = -1
			chrSize = 0
		}
		// end new region test
		//fmt.Println(newval)
		// update histogram
		chrSize = chrSize + 1 // count of data in region
		if uint32(newval) > maxregionhisto {
			maxregionhisto = uint32(newval)
		}
		if newval > histsize-2 {
			histogram[histsize-1] = histogram[histsize-1] + 1
		} else {
			histogram[newval] = histogram[newval] + 1
		}

		// end updat histogram
		// fmt.Println(depthLine[1])
		// fmt.Println(depthLine[2])
		lastRegion = regionName
		if newval == 0 {
			maxzero = maxzero + 1
			zerogram[0]++
		} else if lastval == 0 {
			if maxzero > uint32(zerosize-2) {
				zerogram[zerosize-1] = zerogram[zerosize-1] + 1
				maxzerocon = maxzero
			} else {
				zerogram[maxzero] = zerogram[maxzero] + 1
			}
			maxzero = 0
			for i := 1; i < (zerosize - 1); i++ {
				if zerogram[i] > maxzerogram {
					maxzerogram = zerogram[i]
				}
			}
		}
		lastval = newval
	}
	for i := 1; i < (histsize - 1); i++ {    // could do these in print function (they duplicat of above)
		if histogram[i] > maxhisto {
			maxhisto = histogram[i]
		}
	}
	for i := 1; i < (zerosize - 1); i++ {
		if zerogram[i] > maxzerogram {
			maxzerogram = zerogram[i]
		}
	}
	printRegion(lastRegion, histogram[:], chrSize, maxhisto, maxregionhisto, histsize)
	fmt.Println("><><>< zerogram: ")
	printRegion(lastRegion, zerogram[:], uint64(maxzero), maxzerogram, maxzerocon, zerosize)

	return nil
}

func fileExists(filepath string) bool {
	info, e := os.Stat(filepath)
	if os.IsNotExist(e) {
		return false
	}
	return !info.IsDir()
}

func printRegion(region string, histogram []uint32, chrsize uint64, maxhisto uint32, maxregionhisto uint32, size int) {

	fmt.Println("first 8 and last values of histogram:", region)
	for i := 1; i < 8; i++ {
		fmt.Println("histogram[", i, "] : ", histogram[i])
	}
	fmt.Println("histogram[", size-1, "] : ", histogram[size-1])
	fmt.Println("Number of data read: ", chrsize)
	fmt.Println("Maximum value 1-", size-1, ": ", maxhisto)
	fmt.Println("Maximum value : ", maxregionhisto)
	fmt.Println("Zero count : ", histogram[0])
	// output for plotly in R
	fmt.Println(region, " <- list(")
	fmt.Println("line = list(shape = \"linear\"),")
	fmt.Println("mode = \"lines+markers\",")
	fmt.Println("name = \"'",region,"'\",")
	fmt.Println("type = \"scatter\",")
	fmt.Println("x = c(")
	for i := 1; i < (size-2); i++ {
		fmt.Print(histogram[i],", ")
	}
	fmt.Print(histogram[size-1])
	fmt.Println("),")

	fmt.Println("y = c(")
	for i := 1; i < (size-2); i++ {
		fmt.Print(i,", ")
	}
	fmt.Print(size-1)
	fmt.Println("),")

	fmt.Println("hoverinfo = \"name\"")
	fmt.Println(")")
}
