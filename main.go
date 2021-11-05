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

func main() {
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
	// fmt.Printf("%v\n\n", diff[1:])
	for i, s := range diff[1:] {
		fmt.Printf("%d: %s\n", i, s)
		file1 := strings.Split(s, "|")[0]
		file2 := strings.Split(s, "|")[1]

		filename1 := strings.Split(file1, ":")[0]
		line1 := strings.Split(file1, ":")[1][6:]

		filename2 := strings.Split(file2, ":")[0]
		pos := strings.Index(strings.Split(file2, ":")[1], "[")
		line2 := strings.Split(file2, ":")[1][6:pos]

		beginLine1, _ := strconv.Atoi(strings.Split(line1, "-")[0])
		endLine1, _ := strconv.Atoi(strings.Split(line1, "-")[1])
		beginLine2, _ := strconv.Atoi(strings.Split(line2, "-")[0])
		endLine2, _ := strconv.Atoi(strings.Split(line2, "-")[1])

		fmt.Printf("filename1: %s, from: %d, to: %d\n", filename1, beginLine1, endLine1)
		fmt.Printf("filename2: %s, from: %d, to: %d\n", filename2, beginLine2, endLine2)
		fmt.Println()

		f1, _ := os.Open(filename1)
		f2, _ := os.Open(filename2)

		var code1, code2 string

		scanner := bufio.NewScanner(f1)
		linenum := 1
		fmt.Println(filename1)
		for scanner.Scan() {
			code1 += fmt.Sprintf("%s\n", scanner.Text())
			linenum++
		}
		fmt.Println(code1)

		scanner = bufio.NewScanner(f2)
		linenum = 1
		fmt.Println(filename2)
		for scanner.Scan() {
			code2 += fmt.Sprintf("%s\n", scanner.Text())
			linenum++
		}
		fmt.Println(code1)
		//code1 = html.EscapeString(code1)
		//code2 = html.EscapeString(code2)

		f1.Close()
		f2.Close()

		type templateCodeData struct {
			Title    string
			Code     string
			FromLine int
			ToLine   int
		}

		templateCode, _ := template.ParseFiles("layout_code.html")
		data := templateCodeData{
			Title:    filename1,
			Code:     code1,
			FromLine: beginLine1 - 1,
			ToLine:   endLine1,
		}
		ff, _ := os.Create(filename1 + ".html")
		templateCode.Execute(ff, data)
		ff.Close()

		data = templateCodeData{
			Title:    filename2,
			Code:     code2,
			FromLine: beginLine2 - 1,
			ToLine:   endLine2,
		}
		ff, _ = os.Create(filename2 + ".html")
		templateCode.Execute(ff, data)
		ff.Close()

	}
}
