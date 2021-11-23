package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
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
}

type mapKey struct {
	FileName1, FileName2 string
}

func runSim() []string {
	out, err := exec.Command("bash", "-c", "./sim_c++ -nT -w1 *.cpp").Output()
	if err != nil {
		log.Fatal(err)
	}
	output := string(out)
	lines := strings.Split(output, "\n")
	var diff []string
	for _, s := range lines {
		if s != "" {
			diff = append(diff, s)
		}
	}
	return diff
}

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

func getLines(s string) (int, int) {
	beginLine, _ := strconv.Atoi(strings.Split(s, "-")[0])
	endLine, _ := strconv.Atoi(strings.Split(s, "-")[1])

	return beginLine, endLine
}

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

func genHtml(data templateCodeData) {

	templateCode, _ := template.ParseFiles("layout_code.html")
	// fmt.Println(data)
	f, _ := os.Create(data.FileName1 + "-" + data.FileName2 + ".html")
	templateCode.Execute(f, data)
	f.Close()
}

func main() {
	diff := runSim()

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

		if _, ok := m[mapKey{filename1, filename2}]; !ok {
			m[mapKey{filename1, filename2}] = templateCodeData{
				filename1, filename2, code1, code2,
				[]duplicateLine{},
			}
		}

		temp := m[mapKey{filename1, filename2}].DuplicateLines
		temp = append(temp, duplicateLine{
			LineRange1: lineRange{beginLine1, endLine1},
			LineRange2: lineRange{beginLine2, endLine2},
		})

		m[mapKey{filename1, filename2}].DuplicateLines = temp
	}

	fmt.Println(m)

	for key, _ := range m {
		genHtml(m[key])
	}
}
