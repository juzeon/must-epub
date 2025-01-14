package main

import (
	"bufio"
	"fmt"
	"github.com/bmaupin/go-epub"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type Section struct {
	Title     string
	Content   string
	WordCount int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: mustepub file.txt")
		return
	}
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	title := fileNameWithoutExtSliceNotation(filepath.Base(filePath))
	epubFile := epub.NewEpub(title)
	epubFile.SetLang("zh")
	scanner := bufio.NewScanner(file)
	section := Section{Title: "前言"}
	flushSection := func() {
		_, err := epubFile.AddSection("<h2>"+section.Title+"</h2>"+
			"<p><i>字数："+strconv.Itoa(section.WordCount)+
			"</i></p>"+section.Content, section.Title, "", "")
		if err != nil {
			panic(err)
		}
	}
	for scanner.Scan() {
		line := normalizeLine(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			flushSection()
			section = Section{Title: html.EscapeString(line[2:])}
			fmt.Println("added section: " + line[2:])
		} else {
			section.Content += "<p>" + html.EscapeString(line) + "</p>"
			section.WordCount += countChineseCharacters(line)
		}
	}
	flushSection()
	if err = scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println("writing...")
	err = epubFile.Write(filepath.Join(filepath.Dir(filePath), title+".epub"))
	if err != nil {
		panic(err)
	}
}

var emptyPrefix = regexp.MustCompile(`(?m)^[\s　]+`)
var emptySuffix = regexp.MustCompile(`(?m)[\s　]+$`)

func normalizeLine(line string) string {
	line = emptyPrefix.ReplaceAllString(line, "")
	line = emptySuffix.ReplaceAllString(line, "")
	line = removeNonPrintable(line)
	return line
}
func fileNameWithoutExtSliceNotation(fileName string) string {
	return fileName[:len(fileName)-len(filepath.Ext(fileName))]
}
func countChineseCharacters(s string) int {
	count := 0
	for _, r := range s {
		// 检查字符是否在中文字符的 Unicode 范围内
		// CJK 统一汉字范围: 0x4E00-0x9FFF
		// CJK 扩展 A 区: 0x3400-0x4DBF
		// CJK 扩展 B 区: 0x20000-0x2A6DF
		if (r >= 0x4E00 && r <= 0x9FFF) ||
			(r >= 0x3400 && r <= 0x4DBF) ||
			(r >= 0x20000 && r <= 0x2A6DF) {
			count++
		}
	}
	return count
}
func removeNonPrintable(str string) string {
	// 创建一个 builder 来构建新的字符串
	var builder strings.Builder
	// 遍历原始字符串中的每个字符
	for _, r := range str {
		// 使用 unicode.IsPrint 判断字符是否可打印
		if unicode.IsPrint(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
