package phrase

import "fmt"

// Dict converts between lists of numbers and lists of words.
type Dict struct {
	words []string
	index map[string]uint64
}

// MakeDict creates a Dict from the list of words (vocabulary).
func MakeDict(words []string) (*Dict, error) {
	index := make(map[string]uint64, len(words))
	for i, word := range words {
		if old, has := index[word]; has {
			return nil, fmt.Errorf("word %q occured at indices %d and %d", word, old, i)
		}
		index[word] = uint64(i)
	}
	return &Dict{words, index}, nil
}

// ToWords converts the list of numbers to the list of words.
func (d *Dict) ToWords(numbers []uint64) (words []string, err error) {
	words = make([]string, len(numbers))
	for i, number := range numbers {
		if number >= uint64(len(d.words)) {
			return nil, fmt.Errorf("too high number: %d", number)
		}
		words[i] = d.words[number]
	}
	return
}

// ToNumbers converts the list of words to the list of numbers.
func (d *Dict) ToNumbers(words []string) (numbers []uint64, err error) {
	numbers = make([]uint64, len(words))
	for i, word := range words {
		number, has := d.index[word]
		if !has {
			return nil, fmt.Errorf("unknown word: %q", word)
		}
		numbers[i] = number
	}
	return
}
