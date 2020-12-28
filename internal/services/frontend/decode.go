package frontend

import (
	"errors"
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
	outputkeywords := make([]bool, len(keywords))
	// I need to create putput array of right size with all elements set to False
	
	for _, t := range test{
		t2 := strings.TrimSpace(t)
		for i, k := range keywords{
		
	// Second subtask is to figure out how to test one element in the test slice
	// -- how to tell if test[0] matches any of the keywords elements, and which one.
			if k == t2 {
			//Set the output element of [i] to true 
				outputkeywords[i] = true
			
				//	response = append(outputkeywords[i])

			//else 

			//return error
			
			}
		}
	}
	for _, t3 := range test{
		t4 := strings.TrimSpace(t3)
		found := false
		
		for _, h := range keywords{

	// Output array, nil 
	//How and where we catch the error of mismatch(Write code to just spot the error) write a loop return  an error (true or false)

	// keywords and source are passed into decode and are immutable.
	// 1st subtask creates this: test is created from the data in source
	// 2nd subtack creates this: matches []bool that is created inside decode and 
	// 2nd subtask starts this: ...filled in by the match tests, and then returned
	// 3rd subtask completes this: ...filled in by the match tests, and then returned
			if h == t4{
				found = true
			}
		}
		if found == false {
			return nil, errors.New("Bad view found")
		}
	}
return outputkeywords, nil
}

