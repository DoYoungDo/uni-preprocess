package unipreprocessgo

import (
	"regexp"
	"strconv"
	"strings"
)

type ProcessContext map[string]interface{}

type PreprocessOptions struct {
	Type    string
	Context ProcessContext
}

type PreprocessResult struct {
	Code              string
	IsInPreprocessor func(offset int) bool
}

type preprocessType struct {
	start struct {
		pattern string
		regex   *regexp.Regexp
	}
	end struct {
		pattern string
		regex   *regexp.Regexp
	}
}

var types = map[string]preprocessType{
	"js": {
		start: struct {
			pattern string
			regex   *regexp.Regexp
		}{
			pattern: `[ \t]*(?://|/\*)[ \t]*#(ifndef|ifdef)[ \t]+([^\n*]*)(?:\*(?:\*|/))?(?:[ \t]*\n)?`,
			regex:   regexp.MustCompile(`(?mi)[ \t]*(?://|/\*)[ \t]*#(ifndef|ifdef)[ \t]+([^\n*]*)(?:\*(?:\*|/))?(?:[ \t]*\n)?`),
		},
		end: struct {
			pattern string
			regex   *regexp.Regexp
		}{
			pattern: `[ \t]*(?://|/\*)[ \t]*#endif[ \t]*(?:\*(?:\*|/))?(?:[ \t]*\n)?`,
			regex:   regexp.MustCompile(`(?mi)[ \t]*(?://|/\*)[ \t]*#endif[ \t]*(?:\*(?:\*|/))?(?:[ \t]*\n)?`),
		},
	},
	"html": {
		start: struct {
			pattern string
			regex   *regexp.Regexp
		}{
			pattern: `[ \t]*<!--[ \t]*#(ifndef|ifdef|if)[ \t]+(.*?)[ \t]*(?:-->|!>)(?:[ \t]*\n)?`,
			regex:   regexp.MustCompile(`(?mi)[ \t]*<!--[ \t]*#(ifndef|ifdef|if)[ \t]+(.*?)[ \t]*(?:-->|!>)(?:[ \t]*\n)?`),
		},
		end: struct {
			pattern string
			regex   *regexp.Regexp
		}{
			pattern: `[ \t]*<!(?:--)?[ \t]*#endif[ \t]*(?:-->|!>)(?:[ \t]*\n)?`,
			regex:   regexp.MustCompile(`(?mi)[ \t]*<!(?:--)?[ \t]*#endif[ \t]*(?:-->|!>)(?:[ \t]*\n)?`),
		},
	},
}

type matchGroup struct {
	start int
	end   int
	value string
}

func Preprocess(source string, options PreprocessOptions) PreprocessResult {
	ranges := [][2]int{}
	
	isInPreprocessor := func(offset int) bool {
		for _, r := range ranges {
			if r[0] <= offset && offset < r[1] {
				return true
			}
		}
		return false
	}

	if !strings.Contains(source, "#endif") {
		return PreprocessResult{
			Code:              source,
			IsInPreprocessor: isInPreprocessor,
		}
	}

	context := options.Context
	if context == nil {
		context = make(ProcessContext)
	}

	result := source
	processType := options.Type
	if processType == "" {
		processType = "auto"
	}

	if processType == "auto" || processType == "js" {
		result, ranges = preprocessByType(result, types["js"], context, ranges)
	}
	if processType == "auto" || processType == "html" {
		result, ranges = preprocessByType(result, types["html"], context, ranges)
	}

	return PreprocessResult{
		Code:              result,
		IsInPreprocessor: isInPreprocessor,
	}
}

func preprocessByType(source string, pType preprocessType, context ProcessContext, ranges [][2]int) (string, [][2]int) {
	result := source
	
	for {
		matches := findMatches(result, pType)
		if len(matches) == 0 {
			break
		}
		
		for i := len(matches) - 1; i >= 0; i-- {
			match := matches[i]
			variant := match.variant
			test := strings.TrimSpace(match.test)
			
			switch variant {
			case "ifdef":
				if testPasses(test, context) {
					result = result[:match.startPos] + match.content + result[match.endPos:]
				} else {
					result = result[:match.startPos] + result[match.endPos:]
				}
			case "ifndef":
				if !testPasses(test, context) {
					result = result[:match.startPos] + match.content + result[match.endPos:]
				} else {
					result = result[:match.startPos] + result[match.endPos:]
				}
			}
			ranges = append(ranges, [2]int{match.startPos, match.endPos})
		}
	}
	
	return result, ranges
}

type matchResult struct {
	startPos int
	endPos   int
	content  string
	variant  string
	test     string
}

func findMatches(source string, pType preprocessType) []matchResult {
	var matches []matchResult
	
	startMatches := pType.start.regex.FindAllStringSubmatchIndex(source, -1)
	endMatches := pType.end.regex.FindAllStringIndex(source, -1)
	
	for _, startMatch := range startMatches {
		startPos := startMatch[0]
		startEnd := startMatch[1]
		
		variant := source[startMatch[2]:startMatch[3]]
		test := ""
		if len(startMatch) > 4 && startMatch[4] != -1 {
			test = source[startMatch[4]:startMatch[5]]
		}
		
		for _, endMatch := range endMatches {
			if endMatch[0] > startEnd {
				endPos := endMatch[1]
				content := source[startEnd:endMatch[0]]
				
				matches = append(matches, matchResult{
					startPos: startPos,
					endPos:   endPos,
					content:  content,
					variant:  variant,
					test:     test,
				})
				break
			}
		}
	}
	
	return matches
}

func testPasses(test string, context ProcessContext) bool {
	if test == "" {
		test = "true"
	}
	test = strings.TrimSpace(test)
	test = strings.ReplaceAll(test, "-", "_")
	
	return evaluateExpression(test, context)
}

func evaluateExpression(expr string, context ProcessContext) bool {
	expr = strings.TrimSpace(expr)
	
	if expr == "true" {
		return true
	}
	if expr == "false" {
		return false
	}
	
	if val, exists := context[expr]; exists {
		return isTruthy(val)
	}
	
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return evaluateComparison(left, right, context, "==")
		}
	}
	
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			return evaluateComparison(left, right, context, "!=")
		}
	}
	
	return false
}

func evaluateComparison(left, right string, context ProcessContext, op string) bool {
	leftVal := getValue(left, context)
	rightVal := getValue(right, context)
	
	switch op {
	case "==":
		return leftVal == rightVal
	case "!=":
		return leftVal != rightVal
	}
	return false
}

func getValue(expr string, context ProcessContext) string {
	expr = strings.TrimSpace(expr)
	
	if strings.HasPrefix(expr, `"`) && strings.HasSuffix(expr, `"`) {
		return expr[1 : len(expr)-1]
	}
	if strings.HasPrefix(expr, `'`) && strings.HasSuffix(expr, `'`) {
		return expr[1 : len(expr)-1]
	}
	
	if val, exists := context[expr]; exists {
		return toString(val)
	}
	
	return expr
}

func toString(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func isTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int:
		return v != 0
	case float64:
		return v != 0
	case nil:
		return false
	default:
		return true
	}
}

