package main

import (
	"math/rand"
	"time"
	"bufio"
	"os"
	"fmt"
)

var transforms = make([]string, 0)

func main() {
	fileName := "res/additional_word.txt"
	fp, err := os.Open(fileName)
	defer fp.Close()

	if err != nil {
		fmt.Println("Failed to open " + fileName)
		return
	}

	fpScanner := bufio.NewScanner(fp)
	for fpScanner.Scan() {
		transforms = append(transforms, fpScanner.Text())
	}

	rand.Seed(time.Now().UTC().UnixNano())
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		t := transforms[rand.Intn(len(transforms))]
		if rand.Intn(2) == 0 {
			fmt.Println(t, sc.Text())
		} else {
			fmt.Println(sc.Text(), t)
		}
	}
}
