package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func main() {
	words := make(map[string]bool)

	// load the dictionary in bg so we can do as much preprocessing as we
	// can in the meantime

	wait := make(chan bool)
	go func(ch chan<- bool) {
		// load the preprocessed dictionary that contains only lower
		// case words
		fData, err := ioutil.ReadFile("dictionary.txt")
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
	<-wait

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
