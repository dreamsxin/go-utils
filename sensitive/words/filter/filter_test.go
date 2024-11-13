package filter

import (
	"testing"
)

func TestWordsFilter(t *testing.T) {
	texts := []string{
		"Miyamoto Musashi",
		"妲己",
		"アンジェラ",
		"ความรุ่งโรจน์",
	}
	wf := New()
	root := wf.Generate(texts)
	wf.Remove("shif", root)
	c1 := wf.Contains("アン", root)
	if c1 != false {
		t.Errorf("Test Contains expect false, get %T, %v", c1, c1)
	}
	c2 := wf.Contains("->アンジェラ2333", root)
	if c2 != true {
		t.Errorf("Test Contains expect true, get %T, %v", c2, c2)
	}
	r1 := wf.Replace("Game ความรุ่งโรจน์ i like 妲己 heroMiyamotoMusashi", root)
	if r1 != "Game*************ilike**hero***************" {
		t.Errorf("Test Replace expect Game*************ilike**hero***************,get %T,%v", r1, r1)
	}
}

func TestWordsFilterWithFile(t *testing.T) {
	wf := New()
	// Test generated with file.
	root, _ := wf.GenerateWithFile("./words_test.txt")
	c1 := wf.Contains("妲己，己姓，字妲，為中國商朝最後一位君主帝辛的王后", root)
	if c1 != true {
		t.Errorf("Test Contains expect true, get %T, %v", c1, c1)
	}
}

func TestReplace(t *testing.T) {
	texts := []string{
		"妲己",
	}
	wf := New()
	root := wf.Generate(texts)
	r1 := wf.Replace("妲xxxxx己", root)
	if r1 != "**" {
		t.Errorf("Test Replace expect **,get %T,%v", r1, r1)
	}
	r2 := wf.StrictReplace("妲xxxxx己", root)
	if r2 != "妲xxxxx己" {
		t.Errorf("Test Replace expect 妲xxxxx己,get %T,%v", r2, r2)
	}
}

func TestStrictContains(t *testing.T) {
	texts := []string{
		"妲己",
	}
	wf := New()
	root := wf.Generate(texts)
	c1 := wf.Contains("妲xxxxx己", root)
	if c1 != true {
		t.Errorf("Test Contains expect true, get %T, %v", c1, c1)
	}
	c2 := wf.StrictContains("妲xxxxx己", root)
	if c2 != false {
		t.Errorf("Test Contains expect false, get %T, %v", c2, c2)
	}
}
