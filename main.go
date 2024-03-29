package main

import (
	"bufio"
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/jessevdk/go-flags"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed templates
var templateFile embed.FS

type lineRange struct {
	Begin, End int
}

type duplicateLine struct {
	LineRange1, LineRange2 lineRange
	HighlightColor         int
}

type templateCodeData struct {
	FileName1, FileName2 string
	Code1, Code2         string
	LineNum1, LineNum2   int
	DuplicateLines       []duplicateLine
	DuplicateRate        float64
}

type mapKey struct {
	FileName1, FileName2 string
}

// Run similarity test
func runSim(files []string, opt *Args) []string {
	var output string
	log.Printf("All filenames: %s", files)
	if opt.Language == "" {
		log.Println("No language specified, all language will be tested based on the suffix of the filename.")
		var c_file, cpp_file, java_file, other_file []string
		for _, name := range files {
			if strings.HasSuffix(name, "json") {
				continue
			}
			if strings.HasSuffix(name, "c") {
				c_file = append(c_file, name)
			} else if strings.HasSuffix(name, "cpp") {
				cpp_file = append(cpp_file, name)
			} else if strings.HasSuffix(name, "java") {
				java_file = append(java_file, name)
			} else {
				other_file = append(other_file, name)
			}
		}
		log.Printf("C language files: %s", c_file)
		log.Printf("C++ language files: %s", cpp_file)
		log.Printf("Java language files: %s", java_file)
		log.Printf("Other language files: %s", other_file)
		out, err := exec.Command("bash", "-c", "sim_c -nT -w1 "+strings.Join(c_file, " ")).Output()
		output += string(out)
		out, err = exec.Command("bash", "-c", "sim_c++ -nT -w1 "+strings.Join(cpp_file, " ")).Output()
		output += string(out)
		out, err = exec.Command("bash", "-c", "sim_java -nT -w1 "+strings.Join(java_file, " ")).Output()
		output += string(out)
		out, err = exec.Command("bash", "-c", "sim_text -nT -w1 "+strings.Join(other_file, " ")).Output()
		output += string(out)
		if err != nil {
			log.Fatalln("Run sim_" + opt.Language + " error, please check dependency installation")
		}
	} else {
		out, err := exec.Command("bash", "-c", "sim_"+opt.Language+" -nT -w1 "+strings.Join(files, " ")).Output()
		output = string(out)
		if err != nil {
			log.Fatalln("Run sim_" + opt.Language + " error, please check dependency installation")
		}
	}
	lines := strings.Split(output, "\n")
	var diff []string
	// Get output line by line
	for _, s := range lines {
		// Remove "Total input: 5 files (5 new, 0 old), 249 tokens"
		if s != "" && !strings.HasPrefix(s, "Total input") {
			diff = append(diff, s)
		}
	}
	//log.Println(diff)
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

// Determine whether two files from same person
// Filename format always be like: <id>-<name>_<stu_id>.cpp
// For example: 25-Tom_12345678.cpp
func isSameSource(filename1 string, filename2 string) bool {
	name1 := strings.Split(filename1, "-")
	name2 := strings.Split(filename2, "-")
	if len(name1) < 2 || len(name2) < 2 {
		return false
	}
	if name1[1] == name2[1] {
		return true
	}
	return false
}

// Calculate duplicate rate by lines
func calDuplicateRate(line []duplicateLine, linenum1 int, linenum2 int) float64 {
	var file1lines, file2lines int
	for _, val := range line {
		file1lines += val.LineRange1.End - val.LineRange1.Begin
		file2lines += val.LineRange2.End - val.LineRange2.Begin
	}
	file1lines++
	file2lines++
	file1dup := float64(file1lines) / float64(linenum1)
	file2dup := float64(file2lines) / float64(linenum2)
	if file1dup > file2dup {
		return float64(int(file1dup * 100))
	} else {
		return float64(int(file2dup * 100))
	}
}

// Generate a similar file pair result as a html page
func genHtml(data templateCodeData, opt *Args) {
	templateCode, err := template.ParseFS(templateFile, "templates/layout_code.tmpl")
	if err != nil {
		panic(err)
	}
	// fmt.Println(data)
	err = os.MkdirAll(opt.Output, 0755)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(path.Join(opt.Output, data.FileName1+"-"+data.FileName2+".html"))
	if err != nil {
		panic(err)
	}
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
	templateCode, _ := template.ParseFS(templateFile, "templates/summary.tmpl")
	f, err := os.Create(path.Join(opt.Output, "summary.html"))
	if err != nil {
		panic(err)
	}
	err = templateCode.Execute(f, data)
	if err != nil {
		panic(err)
	}
	f.Close()

	// json file gen
	jsonOutput, err := json.MarshalIndent(data, "", "  ")
	err = os.WriteFile(path.Join(opt.Output, "summary.json"), jsonOutput, 0644)

	// chart file gen
	chartHtml, _ := templateFile.ReadFile("templates/chart.html")
	f, err = os.Create(path.Join(opt.Output, "chart.html"))
	_, err = f.Write(chartHtml)
	chartJS, _ := templateFile.ReadFile("templates/chart.js")
	f, err = os.Create(path.Join(opt.Output, "chart.js"))
	_, err = f.Write(chartJS)
	if err != nil {
		panic(err)
	}
}

type Args struct {
	Language string `short:"l" long:"language" description:"language"`
	Output   string `short:"o" long:"output" description:"output" default:"output"`
}

func main() {
	// Parse parameters
	parser := flags.NewNamedParser("moss_plus", flags.HelpFlag|flags.PassDoubleDash)
	opt := Args{}
	parser.AddGroup("Options", "Options", &opt)
	files, err := parser.Parse()
	if err != nil {
		panic(err)
	}
	diff := runSim(files, &opt)

	m := make(map[mapKey]templateCodeData)

	for i, s := range diff {
		log.Println(s)
		filename1, line1, filename2, line2 := getFiles(s)

		if isSameSource(filename1, filename2) {
			continue
		}

		beginLine1, endLine1 := getLines(line1)
		beginLine2, endLine2 := getLines(line2)

		log.Printf("filename1: %s, from: %d, to: %d\n", filename1, beginLine1, endLine1)
		log.Printf("filename2: %s, from: %d, to: %d\n", filename2, beginLine2, endLine2)

		linenum1, code1 := getCodes(filename1)
		linenum2, code2 := getCodes(filename2)

		_, filename1 = path.Split(filename1)
		_, filename2 = path.Split(filename2)

		if _, ok := m[mapKey{filename1, filename2}]; !ok {
			m[mapKey{filename1, filename2}] = templateCodeData{
				filename1, filename2, code1, code2,
				linenum1, linenum2,
				[]duplicateLine{}, float64(0),
			}
		}

		// Handle many similar line pairs in two files
		t := m[mapKey{filename1, filename2}]
		t.DuplicateLines = append(t.DuplicateLines, duplicateLine{
			LineRange1:     lineRange{beginLine1, endLine1},
			LineRange2:     lineRange{beginLine2, endLine2},
			HighlightColor: i % 5,
		})

		m[mapKey{filename1, filename2}] = t
	}

	// Calculate duplicate rate one by one, rounded to 2 decimals
	for key, _ := range m {
		t := m[key]
		t.DuplicateRate = calDuplicateRate(t.DuplicateLines, t.LineNum1, t.LineNum2)
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
