package frontend

import (
	"errors"
	"strings"
)

var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}

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
			return nil, errors.New("Bad view found")
		}
	}

return outputkeywords, nil
}

