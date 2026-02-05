package unipreprocessgo

import (
	"testing"
)


func TestPreprocess(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Debug mode");
// #endif
console.log("Production code");
// #ifndef RELEASE
console.log("Not release");
// #endif`

	context := ProcessContext{
		"DEBUG":   true,
		"RELEASE": false,
	}

	options := PreprocessOptions{
		Type:    "js",
		Context: context,
	}

	result := Preprocess(source, options)

	expected := `console.log("Debug mode");
console.log("Production code");
console.log("Not release");
`

	if result.Code != expected {
		t.Errorf("Expected length: %d, Got length: %d", len(expected), len(result.Code))
		t.Errorf("Expected bytes: %v", []byte(expected))
		t.Errorf("Got bytes: %v", []byte(result.Code))
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}
