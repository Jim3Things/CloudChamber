package frontend

import (
	"fmt"
	"strings"
)

type errInMatch struct {
	keywords []string
	source string
}

func (e errInMatch) Error() string {
	return fmt.Sprintf(
		"The source %q was not found in keywords %v.",
		e.source,
		e.keywords)
}

// Decode is a function that finds all occurences of the source string in the keyword array.
// - It returns an error if no occurence of the source string is found in the keyword array.
// - When there is no error, the true value returned in boolean array represents that the
// source string is not null and is equal to one or more elements of the keyword array.
func Decode(keywords []string, source string) ([]bool, error){
	var test []string = strings.Split(source, ",")
		outputkeywords := make([]bool, len(keywords))
	
	for _, t := range test{
		t2 := strings.TrimSpace(t)
		found := false
		
		for i, k := range keywords{
			if k == t2 {
				outputkeywords[i] = true
				found = true	
			}
		}
		if found == false {
			return nil, errInMatch{
				keywords: keywords,
				source: source,
			}
		}
	}
	return outputkeywords, nil
}