package main

const constElements = "elements"

var testVarOkay bool

//!PARANO__PRIVATE_TO_FILE
var testVarNotOkay bool

//!PARANO__PRIVATE_TO_FILE
func testFunctionNotOkay() bool {
	var banana = false
	return banana
}

//!PARANO__PRIVATE_TO_FILE
type testType2 struct {
	foo int
}

//!PARANO__PRIVATE_TO_FILE
type testType3 int

// LOCAL PRIVATE STUFF

var localPrivateStuffTest bool
