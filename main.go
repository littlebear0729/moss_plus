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

type duplicateLineList struct {
	BeginLine1, BeginLine2 int
	EndLine1, EndLine2     int
}

type templateCodeData struct {
	Filename1, Filename2 string
	Code1, Code2         string
	DuplicateLines       []duplicateLineList
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

func genHtml(filename1 string, code1 string, filename2 string, code2 string, duplicateLines []duplicateLineList) {

	templateCode, _ := template.ParseFiles("layout_code.html")
	data := templateCodeData{
		Filename1:      filename1,
		Filename2:      filename2,
		Code1:          code1,
		Code2:          code2,
		DuplicateLines: duplicateLines,
	}
	// fmt.Println(data)
	f, _ := os.Create(filename1 + "-" + filename2 + ".html")
	templateCode.Execute(f, data)
	f.Close()
}

func main() {
	diff := runSim()

	for _, s := range diff[1:] {
		fmt.Println(s)
		filename1, line1, filename2, line2 := getFiles(s)
		beginLine1, endLine1 := getLines(line1)
		beginLine2, endLine2 := getLines(line2)

		duplicateLine := []duplicateLineList{
			{
				beginLine1,
				beginLine2,
				endLine1,
				endLine2,
			},
		}

		fmt.Printf("filename1: %s, from: %d, to: %d\n", filename1, beginLine1, endLine1)
		fmt.Printf("filename2: %s, from: %d, to: %d\n", filename2, beginLine2, endLine2)

		_, code1 := getCodes(filename1)
		_, code2 := getCodes(filename2)

		genHtml(filename1, code1, filename2, code2, duplicateLine)
	}
}
