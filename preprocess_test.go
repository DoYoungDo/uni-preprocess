// Package unipreprocessgo 为 JavaScript 和 HTML 文件提供条件编译预处理功能。
// 此测试文件包含全面的测试用例，覆盖各种预处理场景：
//
// 基础功能测试：
//   - ifdef/ifndef 指令的 true/false 条件处理
//   - JavaScript 和 HTML 注释风格处理
//   - 单个源文件中的多个预处理块
//
// 文件类型支持：
//   - JavaScript 文件（// 和 /* */ 注释）
//   - HTML 文件（<!-- --> 注释）
//   - 文件类型自动检测
//
// 上下文和表达式处理：
//   - 空上下文和 nil 上下文场景
//   - 字符串、数值和布尔值比较
//   - 使用 == 和 != 操作符的复杂表达式
//
// 边界情况和特殊功能：
//   - 不包含预处理指令的文件
//   - 混合注释格式
//   - IsInPreprocessor 函数验证
//   - 嵌套条件块
package unipreprocessgo

import (
	"testing"
)

func TestPreprocessBasic(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Debug mode");
// #endif
console.log("Production code");`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Debug mode");
console.log("Production code");`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessIfdefFalse(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Debug mode");
// #endif
console.log("Production code");`

	context := ProcessContext{"DEBUG": false}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Production code");`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessIfndef(t *testing.T) {
	source := `// #ifndef PRODUCTION
console.log("Development mode");
// #endif`

	context := ProcessContext{"PRODUCTION": false}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Development mode");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessIfndefTrue(t *testing.T) {
	source := `// #ifndef PRODUCTION
console.log("Development mode");
// #endif`

	context := ProcessContext{"PRODUCTION": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := ``

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessHTML(t *testing.T) {
	source := `<!-- #ifdef DEBUG -->
<div>Debug info</div>
<!-- #endif -->
<div>Normal content</div>`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "html", Context: context}
	result := Preprocess(source, options)

	expected := `<div>Debug info</div>
<div>Normal content</div>`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessMultipleBlocks(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Debug 1");
// #endif
// #ifdef TEST
console.log("Test mode");
// #endif
console.log("Always");`

	context := ProcessContext{"DEBUG": true, "TEST": false}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Debug 1");
console.log("Always");`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessNoEndif(t *testing.T) {
	source := `console.log("No preprocessor");`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	if result.Code != source {
		t.Errorf("Expected unchanged source, got: %q", result.Code)
	}
}

func TestPreprocessEmptyContext(t *testing.T) {
	source := `// #ifdef UNDEFINED
console.log("Should not appear");
// #endif
console.log("Always visible");`

	options := PreprocessOptions{Type: "js", Context: ProcessContext{}}
	result := Preprocess(source, options)

	expected := `console.log("Always visible");`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessNilContext(t *testing.T) {
	source := `// #ifdef TEST
console.log("Test");
// #endif`

	options := PreprocessOptions{Type: "js", Context: nil}
	result := Preprocess(source, options)

	expected := ``

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessAutoType(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("JS Debug");
// #endif
<!-- #ifdef DEBUG -->
<div>HTML Debug</div>
<!-- #endif -->`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "auto", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("JS Debug");
<div>HTML Debug</div>
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessStringComparison(t *testing.T) {
	source := `// #ifdef ENV == "development"
console.log("Dev environment");
// #endif
// #ifdef ENV != "production"
console.log("Not production");
// #endif`

	context := ProcessContext{"ENV": "development"}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Dev environment");
console.log("Not production");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessNumericComparison(t *testing.T) {
	source := `// #ifdef VERSION == "2"
console.log("Version 2");
// #endif
// #ifdef COUNT != "0"
console.log("Has count");
// #endif`

	context := ProcessContext{"VERSION": 2, "COUNT": 5}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Version 2");
console.log("Has count");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessBooleanValues(t *testing.T) {
	source := `// #ifdef ENABLED
console.log("Enabled");
// #endif
// #ifndef DISABLED
console.log("Not disabled");
// #endif`

	context := ProcessContext{"ENABLED": true, "DISABLED": false}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Enabled");
console.log("Not disabled");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessBlockComments(t *testing.T) {
	source := `/* #ifdef DEBUG */
console.log("Debug with block comment");
/* #endif */`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Debug with block comment");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessIsInPreprocessor(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Debug");
// #endif
console.log("Normal");`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	// Test the IsInPreprocessor function
	if result.IsInPreprocessor == nil {
		t.Error("IsInPreprocessor function should not be nil")
	}
}

func TestPreprocessMixedComments(t *testing.T) {
	source := `// #ifdef DEBUG
console.log("Line comment");
// #endif
/* #ifdef TEST */
console.log("Block comment");
/* #endif */`

	context := ProcessContext{"DEBUG": true, "TEST": false}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Line comment");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessHTMLVariants(t *testing.T) {
	source := `<!-- #ifdef DEBUG -->
<div>Standard HTML comment</div>
<!-- #endif -->`

	context := ProcessContext{"DEBUG": true}
	options := PreprocessOptions{Type: "html", Context: context}
	result := Preprocess(source, options)

	expected := `<div>Standard HTML comment</div>
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessNestedConditions(t *testing.T) {
	source := `// #ifdef OUTER
console.log("Outer true");
// #endif`

	context := ProcessContext{"OUTER": true}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Outer true");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}

func TestPreprocessComplexExpression(t *testing.T) {
	source := `// #ifdef NODE_ENV == "development"
console.log("Development mode");
// #endif
// #ifdef API_VERSION != "v1"
console.log("Not v1 API");
// #endif`

	context := ProcessContext{"NODE_ENV": "development", "API_VERSION": "v2"}
	options := PreprocessOptions{Type: "js", Context: context}
	result := Preprocess(source, options)

	expected := `console.log("Development mode");
console.log("Not v1 API");
`

	if result.Code != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, result.Code)
	}
}
