package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

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
	var histsize = 1024
	var histogram [1024]uint32 // depth histogram for whole chromosome 1 - 511 no 0 as this will be to big
	var maxhisto uint32       // region max of histo from 1 to 510
	var maxzero uint32
	var maxzerogram uint32
	var maxzerocon uint32
	const zerosize = 750
	var zerogram [zerosize]uint32
	maxhisto = 0

	var maxregionhisto uint32
	//	var regionpercenthistogram[512] uint // region % histogram
	var regionName string
	var lastRegion string
	lastRegion = "_last_region"

	// output file setup
	t := time.Now()
	ft := t.Format("060102T1504") + "_depthHist.R"
	afile, err := os.Create(ft)
	check(err)
	defer afile.Close()
	// write header
	l, err := afile.WriteString("p <- plot_ly()\n")
	if err != nil {
		fmt.Println(err, " Error writing depth header to ", ft)
	} else {
		fmt.Println(l, "bytes of Header written to ", ft)
	}
	// end file setup

		// output file zreogram setup
		dft := t.Format("060102T1504") + "_zeroHist.R"
		zfile, err := os.Create(dft)
		check(err)
		defer zfile.Close()
		// write header
		l, err = zfile.WriteString("p <- plot_ly()\n")
		if err != nil {
			fmt.Println(err, " Error writing zero header to ", dft)
		} else {
			fmt.Println(l, "bytes of Header written to ", dft)
		}
		// end file setup

	for {
		// text expected "chr1  1   1" region, position , readdepth
		// (if -a specified in samtools depth the position will be 0- length of region) otherwise 0 depths are omitted from output)

		// depthLine = strings.Fields(scanner.Text())
		line, err := (scanner.ReadString(10)) // read up to eol

		if err == io.EOF {
			fmt.Println("finshed reading input") // if EOF is at end of line last line will be skipped
			break                                // if EOF is on  a newline  then this is ok. could add check for line and have another exit at end of loop?
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
			for i := 1; i < (histsize - 1); i++ { // find max of histogram excluding first and last array spots
				if histogram[i] > maxhisto {
					maxhisto = histogram[i]
				}
			}
			printRegion(lastRegion, histogram[:], chrSize, maxhisto, maxregionhisto, histsize, afile, true)
			maxhisto = 0
			maxregionhisto = 0

			for i := range histogram {
				histogram[i] = 0
			}
			fmt.Println("><><>< zerogram: ")
			printRegion(lastRegion, zerogram[:], chrSize, maxzerogram, maxzerocon, zerosize, zfile, true)
			for i := range zerogram {
				zerogram[i] = 0
			}
			maxzero = 0
			maxzerocon = 0
			maxzerogram = 0
			lastval = -1
			chrSize = 0

			if regionName == "chrM" {
				break  // only get the primary chromosomes
			}

		}
		// end new region test
		//fmt.Println(newval)
		// update histogram
		chrSize = chrSize + 1 // count of data in region
		if uint32(newval) > maxregionhisto {
			maxregionhisto = uint32(newval) // maximum depth in current region
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
	for i := 1; i < (histsize - 1); i++ { // could do these in print function (they duplicat of above)
		if histogram[i] > maxhisto {
			maxhisto = histogram[i]
		}
	}
	for i := 1; i < (zerosize - 1); i++ {
		if zerogram[i] > maxzerogram {
			maxzerogram = zerogram[i]
		}
	}
	/* printRegion(lastRegion, histogram[:], chrSize, maxhisto, maxregionhisto, histsize, afile, true)
	fmt.Println("><><>< zerogram: ")
	printRegion(lastRegion, zerogram[:], uint64(maxzero), maxzerogram, maxzerocon, zerosize, zfile, true)
 */
	l, err = afile.WriteString("\n\n p <- layout(p, xaxis = list(type = \"log\"), yaxis = list(type = \"log\")) ")
	l, err = afile.WriteString("\n\np\n")

	l, err = zfile.WriteString("\n\n p <- layout(p, xaxis = list(type = \"log\"), yaxis = list(type = \"log\")) ")
	l, err = zfile.WriteString("\n\np\n")
	return nil
}

func fileExists(filepath string) bool {
	info, e := os.Stat(filepath)
	if os.IsNotExist(e) {
		return false
	}
	return !info.IsDir()
}

func printRegion(region string, histogram []uint32, chrsize uint64, maxhisto uint32, maxregionhisto uint32, size int, afile *os.File, rout bool) {
	// var xp float64 // percentage holder
	var sxp string
	var dpthMax int
	fmt.Println("first 8 and last values of histogram:", region)
	for i := 1; i < 8; i++ {
		fmt.Println("histogram[", i, "] : ", histogram[i])
	}
	fmt.Println("histogram[", size-1, "] : ", histogram[size-1])
	fmt.Println("Number of data read: ", chrsize)
	maxhisto = 0
	for i := 1; i < (size - 1); i++ { // find max of histogram excluding first and last array spots
		if histogram[i] > maxhisto {
			maxhisto = histogram[i]
			dpthMax = i
		}
	}
	fmt.Println("Maximum depth count 1-", size-1, ": ", maxhisto," corresponding to read depth of ", dpthMax)
	fmt.Println("Maximum depth found : ", maxregionhisto)
	fmt.Println("Zero count : ", histogram[0])

	// output for plotly in R
	//debug
/* 	fmt.Println("only first ten values printed see file for more.\n", region, " <- list(")
	fmt.Print("line = list(shape = \"spline\", width = 1), marker = list(size=3), ")
	fmt.Print("mode = \"lines+markers\",")
	fmt.Print("name = \"", region, "\",")
	fmt.Print("type = \"scatter\",")
	fmt.Print("y = c(") */

	rString := "\n\n" + region + " <- list("
	rString += "line = list(shape = \"spline\", width = 1), marker = list(size=3), "
	rString += "mode = \"lines+markers\","
	rString += "name = \"" + region + "\","
	rString += "type = \"scatter\",\n"

	sString := "y = c("
	var chruniqCnt uint32
	for i := 1; i <= (size - 2); i++ {
		chruniqCnt += histogram[i]
	}
	for i := 1; i <= (size - 2); i++ {
		if chrsize != 0 { // can be zero for the zero cnt case
			// get %
/* 			xp = 100.0 * float64(histogram[i]) / float64(chruniqCnt)
			// fmt.Println(xp)   // test
			// fmt.Printf("%.2f, (",xp)
			xp = float64(histogram[i]) / float64(chruniqCnt)
			if i == (size - 2) {
				sxp = fmt.Sprintf("%.6f", xp) // no comma after last value!
			} else {
				sxp = fmt.Sprintf("%.6f, ", xp)
			} */
// Using histogram uint value in ahisto_z
			sxp = fmt.Sprintf("%d,", histogram[i])

			sString = sString + sxp
			if (i % 100) == 0 { // R does not like long lines (prob in thousands?)
				sString = sString + "\n"
			}
		} else {
			sString += strconv.FormatUint(uint64(histogram[i]), 10)
			if i < (size - 2) { // add comma if not last
				sString += ", "
			}
		}

		// printing to std out only do first 10
		// debug R gen
/* 		if chrsize != 0 { // can be zero for the zero cnt case
			if i < 10 {
				fmt.Print((uint64(histogram[i])*100)/chrsize, ", ") // % of read depth
			}
		} else {
			if i < 10 {
				fmt.Print(histogram[i], ", ") // % of read depth
			}
		} 
	fmt.Println("),")
	fmt.Print("x = c(") */

	}  // end histogram loop

	// DONT FORGET THE EXTRA VALUE!
	// fmt.Print(histogram[size-2])  // omit last value of array which stores count of values above size
	// xp = float64(histogram[size-2]) / float64(chrsize)
	// sxp = fmt.Sprintf("%.2f, ",xp)
	// sString = sString + sxp
	//sString += strconv.FormatUint(uint64(histogram[size-2]),10)

	sString += "),\n"


	sString += "x = c("
	for i := 1; i < (size - 2); i++ {
		if i < 10 {
			fmt.Print(i, ", ")
		}
		sString += strconv.Itoa(i) + ", "
		if (i % 100) == 0 { // R does not like long lines (prob in thousands?)
			sString = sString + "\n"
		}
	}

	sString += strconv.Itoa(size - 2)
	sString += "),\n"

	// debug R
/* 	fmt.Print(size - 2)
	fmt.Println("),")
	fmt.Print(")\n") */

	aStr := "hovertemplate = \"Depth: %{x}\""
	aStr += ")\n"
	aStr += "p <- add_trace"
	aStr += "(p, line=" + region + "$line, mode=" + region + "$mode, "
	aStr += "name=" + region + "$name, type=" + region + "$type, "
	aStr += "marker=" + region + "$marker, "
	aStr += "x=" + region + "$x, y=" + region + "$y, " + "text=" + region + "$text, "
	aStr += "hovertemplate=" + region + "$hovertemplate)"

	// debug R	
/* 	fmt.Println(aStr) */

	if rout {
		_, err := afile.WriteString(rString)
		_, err = afile.WriteString(sString)
		_, err = afile.WriteString(aStr)
		if err != nil {
			fmt.Println(err, " Error writing region ", region, " to ", afile.Name())
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
