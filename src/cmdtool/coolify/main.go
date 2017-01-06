package main

import (
	"math/rand"
	"time"
	"bufio"
	"os"
	"fmt"
)

const (
	duplicateVowel bool = true
	removeVowel bool = false
)

func randBool() bool {
	return rand.Intn(2) == 0
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	s := bufio.NewScanner(os.Stdin)

	for s.Scan() {
		word := []byte(s.Text())
		if randBool() {
			var vI int = -1
			for i, char := range word {
				switch char {
				case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
					if randBool() {
						vI = i
					}
				}
			}
			if vI >= 0 {
				switch randBool() {
				case duplicateVowel:
					// 1文字重複させて増やす
					word = append(word[:vI + 1], word[vI:]...)
				case removeVowel:
					// 1文字飛ばして減らす
					word = append(word[:vI], word[vI + 1:]...)
				}
			}
		}
		fmt.Println(string(word))
	}
}