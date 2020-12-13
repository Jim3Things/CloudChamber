package frontend

import (
	"strings"
)

var keywords []string = []string  {"Defined", "Actual", "Target", "Observed"}

// Input and output for a success case
var views = "Defined, Observed"
var response []string
var expected []bool = []bool {true, false, false, true}

// Input and output for a faiure case
var badView = "AbaddefinedobservedValue" 
var badexpected []bool = nil

func decode(keywords []string, source string) ([]bool, error){
	// First subtask is to get from source to test (convert the source string into the test string slice)
	// hint: strings package probably has something like split and trim.

	var test []string = strings.Split(source, ",")
	
	// Third subtask is to loop through all the test slice elements.
	
	for i := 0; i < 3; i++ {

	// Second subtask is to figure out how to test one element in the test slice
	// -- how to tell if test[0] matches any of the keywords elements, and which one.
		if (strings.Contains(test[i], keywords[i])) {

			 response = append(response, keywords[i])
			
		}
		
	} 
	
	// keywords and source are passed into decode and are immutable.
	// 1st subtask creates this: test is created from the data in source
	// 2nd subtack creates this: matches []bool that is created inside decode and 
	// 2nd subtask starts this: ...filled in by the match tests, and then returned
	// 3rd subtask completes this: ...filled in by the match tests, and then returned

	}