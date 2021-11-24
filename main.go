package main

import (
	"bufio"
	"fmt"
	"github.com/jessevdk/go-flags"
	"html/template"
	"log"
	"math"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

type lineRange struct {
	Begin, End int
}

type duplicateLine struct {
	LineRange1, LineRange2 lineRange
}

type templateCodeData struct {
	FileName1, FileName2 string
	Code1, Code2         string
	DuplicateLines       []duplicateLine
	DuplicateRate        float64
}

type mapKey struct {
	FileName1, FileName2 string
}

// Run similarity test
func runSim(files []string, opt *Args) []string {
	out, err := exec.Command("bash", "-c", "sim_"+opt.Language+" -nT -w1 "+strings.Join(files, " ")).Output()
	if err != nil {
		log.Fatal(err)
	}
	output := string(out)
	lines := strings.Split(output, "\n")
	var diff []string
	// Get output line by line
	for _, s := range lines {
		if s != "" {
			diff = append(diff, s)
		}
	}
	return diff
}

// Get filename and line pairs in an output line
func getFiles(f string) (string, string, string, string) {
	file1 := strings.Split(f, "|")[0]
	file2 := strings.Split(f, "|")[1]

	filename1 := strings.Split(file1, ":")[0]
	line1 := strings.Split(file1, ":")[1][6:]

	filename2 := strings.Split(file2, ":")[0]
	pos := strings.Index(strings.Split(file2, ":")[1], "[")
	line2 := strings.Split(file2, ":")[1][6:pos]

	return filename1, line1, filename2, line2
}

// Get begin and end line in a line pair
func getLines(s string) (int, int) {
	beginLine, _ := strconv.Atoi(strings.Split(s, "-")[0])
	endLine, _ := strconv.Atoi(strings.Split(s, "-")[1])

	return beginLine, endLine
}

// Get code line by line from a filename
func getCodes(filename string) (int, string) {
	f, _ := os.Open(filename)
	var code string
	scanner := bufio.NewScanner(f)
	linenum := 0
	fmt.Println(filename)
	for scanner.Scan() {
		code += fmt.Sprintf("%s\n", scanner.Text())
		linenum++
	}
	fmt.Println(code)
	return linenum, code
}

// Calculate duplicate rate by lines
func calDuplicateRate(line []duplicateLine) float64 {
	var file1lines, file2lines int
	for _, val := range line {
		file1lines += val.LineRange1.End - val.LineRange1.Begin
		file2lines += val.LineRange2.End - val.LineRange2.Begin
	}
	if file1lines >= file2lines {
		return math.Round(float64(file2lines)/float64(file1lines)*100) / 100
	} else {
		return math.Round(float64(file1lines)/float64(file2lines)*100) / 100
	}
}

// Generate a similar file pair result as a html page
func genHtml(data templateCodeData, opt *Args) {

	templateCode, _ := template.ParseFiles("layout_code.tmpl")
	// fmt.Println(data)
	err := os.MkdirAll(opt.Output, 0755)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(path.Join(opt.Output, data.FileName1+"-"+data.FileName2+".html"))
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	err = templateCode.Execute(f, data)
	if err != nil {
		panic(err)
	}
	f.Close()
}

// Generate all similar file pairs result as a html page
func genSummary(data []templateCodeData, opt *Args) {
	sort.Slice(data, func(i, j int) bool {
		return data[i].DuplicateRate > data[j].DuplicateRate
	})

	templateCode, _ := template.ParseFiles("summary.tmpl")
	f, err := os.Create(path.Join(opt.Output, "summary.html"))
	if err != nil {
		panic(err)
	}
	err = templateCode.Execute(f, data)
	if err != nil {
		panic(err)
	}
	f.Close()
}

type Args struct {
	Language string `short:"l" long:"language" description:"language" default:"c++"`
	Output   string `short:"o" long:"output" description:"output" default:"output"`
}

func main() {
	// Parse parameters
	parser := flags.NewNamedParser("eduOJ server", flags.HelpFlag|flags.PassDoubleDash)
	opt := Args{}
	parser.AddGroup("Options", "Options", &opt)
	files, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	diff := runSim(files, &opt)

	m := make(map[mapKey]templateCodeData)

	for _, s := range diff[1:] {
		fmt.Println(s)
		filename1, line1, filename2, line2 := getFiles(s)
		beginLine1, endLine1 := getLines(line1)
		beginLine2, endLine2 := getLines(line2)

		fmt.Printf("filename1: %s, from: %d, to: %d\n", filename1, beginLine1, endLine1)
		fmt.Printf("filename2: %s, from: %d, to: %d\n", filename2, beginLine2, endLine2)

		_, code1 := getCodes(filename1)
		_, code2 := getCodes(filename2)

		_, filename1 = path.Split(filename1)
		_, filename2 = path.Split(filename2)

		if _, ok := m[mapKey{filename1, filename2}]; !ok {
			m[mapKey{filename1, filename2}] = templateCodeData{
				filename1, filename2, code1, code2,
				[]duplicateLine{}, float64(0),
			}
		}

		// Handle many similar line pairs in two files
		t := m[mapKey{filename1, filename2}]
		t.DuplicateLines = append(t.DuplicateLines, duplicateLine{
			LineRange1: lineRange{beginLine1, endLine1},
			LineRange2: lineRange{beginLine2, endLine2},
		})

		m[mapKey{filename1, filename2}] = t
	}

	// Calculate duplicate rate one by one, rounded to 2 decimals
	for key, _ := range m {
		t := m[key]
		dupRate := calDuplicateRate(t.DuplicateLines)
		t.DuplicateRate = dupRate * 100
		m[key] = t
	}

	// Prepare for generating summary page
	var allDup []templateCodeData
	for key, _ := range m {
		genHtml(m[key], &opt)
		allDup = append(allDup, m[key])
	}

	genSummary(allDup, &opt)
}
