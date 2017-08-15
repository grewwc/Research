package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/astrogo/fitsio"
)

func sum(a []float64) float64 {
	if len(a) == 0 {
		return 0
	}
	s := 0.0
	for _, v := range a {
		s += v
	}
	return s
}

func isIn(e string, l []string) bool {
	res := false
	for _, v := range l {
		if e == v {
			res = true
			break
		}
	}
	return res
}

func Tolower(l []string) []string {
	var res []string
	for _, v := range l {
		res = append(res, strings.ToLower(v))
	}
	return res
}

func findMin(a []float64) float64 {
	min := math.MaxFloat64
	for _, v := range a {
		if v < min {
			min = v
		}
	}
	return min
}

func findMax(a []float64) float64 {
	max := .0
	for _, v := range a {
		if v > max {
			max = v
		}
	}
	return max
}

func main() {
	//Parameters that can be passed in
	filePtr := flag.String("f", "", "fits file name")
	listTable := flag.Bool("lt", false, "list all the tables in the fits file")
	showBasic := flag.Bool("show", false, "show the basic information of the fits table")
	printNumOfCol := flag.Bool("nc", false, "column number of events table")
	printNumOfRow := flag.Bool("nr", false, "row number of events table")
	showMinMax := flag.String("range", "", "find the mix and max of particular column in Events table")
	// allArguments := flag.Args()
	flag.Parse()
	//fmt.Println(len(allArguments))

	fname := *filePtr
	colOfRange := ""
	if *showMinMax != "" {
		colOfRange = strings.ToUpper(*showMinMax)
	}
	file, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	hdu, err := fitsio.Open(file)
	if err != nil {
		log.Fatal(err)
	}

	numberOfExtensions := len(hdu.HDUs())

	allHdus := make([]fitsio.HDU, 1)

	for _, v := range hdu.HDUs() {
		allHdus = append(allHdus, v)
	}

	show := func() {
		fmt.Println("\n\n")
		fmt.Println(strings.Repeat("==", 55))
		fmt.Println(strings.Repeat("==", 55))
		fmt.Println()

		fmt.Println("\n")
		for i := 0; i != numberOfExtensions; i++ {
			fmt.Println(hdu.HDU(i).Name(), "(Headers) ", ":")
			for _, v := range hdu.HDU(i).Header().Keys() {
				fmt.Printf("\t")
				fmt.Println(v, " : ", *hdu.HDU(i).Header().Get(v))
			}
			fmt.Println()
		}

		for i := 0; i != numberOfExtensions; i++ {
			fmt.Print(strings.Repeat(" ", 4))
			fmt.Printf("%d: %s\n", i+1, hdu.HDU(i).Name())
		}
	}

	nc := func() {
		if !hdu.Has("EVENTS") {
			log.Fatalln("Don't have events table")
			return
		}
		event := hdu.Get("EVENTS")
		fmt.Print("Number of Columns in Events Table:")
		fmt.Println("   ", event.(*fitsio.Table).NumCols())
	}

	nr := func() {
		if !hdu.Has("EVENTS") {
			log.Fatalln("Don't have events file")
			return
		}
		event := hdu.Get("EVENTS")
		fmt.Print("Number of Rows in Events Table:")
		fmt.Println("   ", event.(*fitsio.Table).NumRows())
	}

	lt := func() {
		num := 0
		fmt.Println()
		for _, v := range allHdus {
			if table, ok := v.(*fitsio.Table); ok == true {
				fmt.Println("#", table.Name(), ":")
				num++
				for _, v1 := range table.Cols() {
					fmt.Print("   ", "\"", v1.Name, "\"")
				}
			} else {
				fmt.Errorf("cannot transfer to table")
				continue
			}
			fmt.Println("\n")
		}
		if num == 0 {
			fmt.Println("\tNo extensions can be transfered to Tables\n")
		}
	}

	minAndmax := func() {
		if !hdu.Has("EVENTS") {
			log.Fatalln("Don't have events table")
			return
		}
		event := hdu.Get("EVENTS")
		table := event.(*fitsio.Table)
		// allCols := table.Cols()
		ncols, nrows := table.NumCols(), table.NumRows()
		step := nrows / 5.0
		hasCol := false
		func() {
			for i := 0; i != ncols; i++ {
				if table.Col(i).Name == colOfRange {
					hasCol = true
					return
				}
			}
		}()
		if hasCol == false {
			fmt.Println("Events file don't have the column name")
			return
		}
		allRows := [5](*fitsio.Rows){}
		for i := 0; i != 5; i++ {
			allRows[i], err = table.Read(int64(i)*int64(step), int64(i+1)*int64(step))
			if err != nil {
				log.Fatal(err)
			}
		}
		wg := sync.WaitGroup{}
		wg.Add(5)
		max := []float64{.0, .0, .0, .0, .0}
		min := []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
		run := func(ith int) {
			name := make([]string, ncols)
			data := make([]interface{}, ncols)
			for i := 0; i != ncols; i++ {
				data[i] = reflect.New(table.Col(i).Type()).Interface()
				name[i] = table.Col(i).Name
			}
			for irow := 0; allRows[ith].Next(); irow++ {
				if err := allRows[ith].Scan(data...); err != nil {
					log.Fatal(err)
				}
				i := 0
				for i = 0; i != ncols; i++ {
					if name[i] == colOfRange {
						target := reflect.Indirect(reflect.ValueOf(data[i])).Float()
						if target > max[ith] {
							max[ith] = target
						}
						if target < min[ith] {
							min[ith] = target
						}
						break
					}
				}
			}
			wg.Done()
		}

		for i := 0; i != 5; i++ {
			go run(i)
		}
		wg.Wait()
		minf := findMin(min)
		maxf := findMax(max)
		fmt.Println("\t", colOfRange+"min", minf)
		fmt.Println("\t", colOfRange+"max", maxf)

	}

	//belows are choices that if function will be needed.
	if *showBasic == true {
		show()
	}

	if *printNumOfCol {
		nc()
	}

	if *printNumOfRow == true {
		nr()
	}

	if *listTable {
		lt()
	}
	/* events := hdu.HDU(1)

	table := events.(*fitsio.Table)
	fmt.Println(table.NumCols()) */
	if colOfRange != "" {
		minAndmax()
	}

}

/* func split(n int) {
	start, stop := 283996802.0, 500000000.0
} */
