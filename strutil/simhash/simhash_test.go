package simhash

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func simhashString(s string) uint64 {
	scanner := bufio.NewScanner(strings.NewReader(s))
	scanner.Split(ScanByteTrigrams)

	return Hash(scanner)
}

func simhashString2(s string) uint64 {
	var r = regexp.MustCompile(`[\w']+(?:\://[\w\./]+){0,1}`)
	words := r.FindAll([]byte(s), -1)

	fmt.Println("words", words)
	return Hash(NewSliceScanner(words))
}

func TestSimSipHash(t *testing.T) {

	h1 := simhashString("Now is the winter of our discontent and also the time for all good people to come to the aid of the party")
	fmt.Printf("h=%016x\n", h1)

	h2 := simhashString("Now is the winter of our discontent and also the time for all good people to come to the party")
	fmt.Printf("h=%016x\n", h2)

	h3 := simhashString("The more we get together together together the more we get together the happier we'll be")
	fmt.Printf("h=%016x\n", h3)

	fmt.Printf("d(h1,h2)=%d\n", Distance(h1, h2))
	fmt.Printf("d(h1,h3)=%d\n", Distance(h1, h3))
	fmt.Printf("d(h2,h3)=%d\n", Distance(h2, h3))

	h4 := simhashString(strings.Repeat("Now is the winter", 241)) // length = 4097
	fmt.Printf("h=%016x\n", h4)

	h5 := simhashString2("this is a test phrase")
	fmt.Printf("h5=%016x\n", h5)

	h6 := simhashString2("this is a test phrass")
	fmt.Printf("h6=%016x\n", h6)
}

func TestSimHash(t *testing.T) {
	var docs = [][]byte{
		[]byte("this is a test phrase"),
		[]byte("this is a test phrass"),
		[]byte("foo bar"),
	}

	hashes := make([]uint64, len(docs))
	for i, d := range docs {
		hashes[i] = Simhash(NewWordFeatureSet(d))
		fmt.Printf("Simhash of %s: %x\n", d, hashes[i])
	}

	fmt.Printf("Comparison of `%s` and `%s`: %d\n", docs[0], docs[1], Compare(hashes[0], hashes[1]))
	fmt.Printf("Comparison of `%s` and `%s`: %d\n", docs[0], docs[2], Compare(hashes[0], hashes[2]))
}
