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
	return fi.Mode() & os.ModeCharDevice == 0
}

func getFile() (*os.File, error){
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
	var chrSize uint64
	var histsize = 512
	var histogram[512] uint32  // depth histogram for whole chromosome 1 - 511 no 0 as this will be to big
	var maxhisto uint32 // region max of histo from 1 to 510
	maxhisto = 0
//	var regionhistogram[512] uint32 // depth histogram for each region
	var maxregionhisto uint32
//	var regionpercenthistogram[512] uint // region % histogram
//	var regionzero[1024] uint32  // region zero run counts
//	var regionmax int32 // max depth per region, ie greater than 511
	var regionName string
	var lastRegion string
	lastRegion = "_last_region" 

	
	for {
		// text expected "chr1  1   1" region, position , readdepth
		// (if -a specified in samtools depth the position will be 0- length of region) otherwise 0 depths are omitted from output)

		// depthLine = strings.Fields(scanner.Text())
		line, err := (scanner.ReadString(10))  // read up to eol
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
		case len(depthLine)	== 0:
			continue  // empty line, go read next line ?
		default	:
			newval = -1
		}
		if err != nil {
			fmt.Println(" data non numeric/integer, exiting")	
			os.Exit(-1)	
		}
		// end get data from file line
		// now test if we have new region
		if regionName != lastRegion {
			for i := 1; i < 511; i++ {
				if histogram[i] > maxhisto {
					maxhisto = histogram[i]
				}
			}
			printRegion(lastRegion,histogram[:],chrSize, maxhisto, maxregionhisto)
			maxhisto = 0
			maxregionhisto = 0
		}
		// end new region test 
		//fmt.Println(newval)
		// update histogram
		chrSize = chrSize + 1  // count of data in region
		if uint32(newval) > maxregionhisto {
			maxregionhisto = uint32(newval)
		}
		if newval > histsize - 1 {
			histogram[histsize-1] = histogram[histsize-1] + 1
		} else {
			histogram[newval] = histogram[newval] + 1
		}

		// end updat histogram
		// fmt.Println(depthLine[1])
		// fmt.Println(depthLine[2])
		lastRegion = regionName
		}
		for i := 1; i < 511; i++ {
			if histogram[i] > maxhisto {
				maxhisto = histogram[i]
			}	
		}		
		printRegion(lastRegion,histogram[:],chrSize,maxhisto,maxregionhisto)
	return nil
}

func fileExists(filepath string) bool {
	info, e := os.Stat(filepath)
	if os.IsNotExist(e) {
		return false
	}
	return !info.IsDir()
}

func printRegion(region string, histogram []uint32, chrsize uint64, maxhisto uint32, maxregionhisto uint32  ) {

	fmt.Println("first 8 and last values of histogram:",region)
	for i := 0; i < 8; i++ {
		fmt.Println("histogram[",i,"] : ",histogram[i])
	}
	fmt.Println("histogram[511] : ",histogram[511])
	fmt.Println("Number of data read: ",chrsize)
	fmt.Println("Maximum value 1-510: ",maxhisto)
	fmt.Println("Maximum value : ",maxregionhisto)
	fmt.Println("Zero count : ",histogram[0])
}
