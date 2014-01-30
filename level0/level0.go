package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	words := make(map[string]bool)

	wait := make(chan bool)
	go func(ch chan <- bool) {
		fData, err := ioutil.ReadFile("preprocessed.txt")
		if err != nil {
			log.Fatalln(err)
			return
		}

		for _, word := range strings.Split(string(fData), "\n") {
			words[word] = true
		}

		ch <- true
	}(wait)

	content, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalln(err)
		return
	}

	_lines := strings.Split(string(content), "\n")
	lines := make([][]string, len(_lines))
	for l, line := range _lines {
		lines[l] = strings.Split(line, " ")
	}
	output := make([]string, len(lines))
	<- wait

	for l, line := range lines {
		for i, word := range line {
			if _, ok := words[strings.ToLower(word)]; !ok {
				lines[l][i] = "<" + word + ">"
			}
		}
		output[l] = strings.Join(lines[l], " ")
	}

	str := strings.Join(output, "\n")
	fmt.Println(str[:len(str)-1])
}
