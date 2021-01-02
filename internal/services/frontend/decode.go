package frontend

import (
	"errors"
	"strings"
)

// ErrInMatch is a new error declaration for the instance if a bad view is found
var errInMatch  = errors.New("decode: Bad view found")

// function to decode the input array in comparison to the source array
func decode(keywords []string, source string) ([]bool, error){
	
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
			return nil, errInMatch
		}
	}

return outputkeywords, nil
}