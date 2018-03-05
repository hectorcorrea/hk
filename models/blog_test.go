package models

import (
	"regexp"
	"strings"
	"testing"
)

func TestSlug(t *testing.T) {
	testA := []string{"", ""}
	testB := []string{"abc 345 DEF", "abc-345-def"}
	testC := []string{"hello c#", "hello-c-sharp"}
	testD := []string{"a<b", "a-b"}
	testE := []string{"a <  b", "a-b"}
	testF := []string{"a b<", "a-b"}
	testG := []string{"a b<<", "a-b"}
	testH := []string{"<", ""}
	tests := [][]string{testA, testB, testC, testD, testE, testF,
		testG, testH}
	for _, test := range tests {
		value := test[0]
		slug := getSlug(value)
		expected := test[1]
		if slug != expected {
			t.Errorf("Unexpected slug (%s) for (%s)", slug, value)
		}
	}
}

// func TestLegacyViewPicture(t *testing.T) {
// 	part1 := "before1 <a href=aaa>bbb<img src=ccc /></a> after1"
// 	part2 := "before2 <a href=xxx>yyy<img src=zzz /></a> after2"
// 	testHtml := part1 + part2
// 	// testHtml = "before <a href=\"javascript:viewpicture('Paris+2008','pictures/2008/halloween_043.jpg')\"><img border=\"1\" alt=\"\" src=\"http://www.hectorykarla.com/photos/2008/halloween_043_thumb.jpg\" /></a> after"
// 	reViewPicture := regexp.MustCompile("<a href=(.*?)>(.*?)</a>")
// 	matches := reViewPicture.FindAllString(testHtml, -1)
//
// 	// match
// 	// <a href=viewpic(...)><img src=yyy /></a>
// 	// 0123456789-123456789-123456789-123456789
//
// 	for i, match := range matches {
// 		img := ""
// 		jsViewPicture := strings.Index(match, "javascript:viewpicture")
// 		if jsViewPicture != -1 {
// 			imgBegin := strings.Index(match, "<img")
// 			if imgBegin > jsViewPicture {
// 				imgEnd := strings.Index(match[imgBegin:], "/>")
// 				if imgEnd > imgBegin {
// 					// a := imgBegin
// 					// b := imgBegin + imgEnd + 2
// 					// t.Errorf("%d %d %d", i, a, b)
// 					img = match[imgBegin:(imgBegin + imgEnd + 2)]
// 				}
// 			}
// 		}
// 		t.Errorf("%d %s", i, img)
// 		if img != "" {
// 			testHtml = strings.Replace(testHtml, match, img, 1)
// 		}
// 	}
//
// 	// str := x[0] //+ " ** " + x[1]
// 	t.Errorf("%s", testHtml)
// }
