package convert

import (
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	reporters "github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

//go:embed style.css
var styles string

var output string

func Convert(suites *reporters.JUnitTestSuites) (string, error) {
	output += "<html>"
	output += "<head>"
	output += "<meta charset=\"UTF-8\">"
	output += "<style>" + styles + "</style>"
	output += "</head>"
	output += "<body>"

	failures, total := 0, 0
	for _, s := range suites.TestSuites {
		failures += s.Failures + s.Errors
		total += len(s.TestCases)
	}
	output += fmt.Sprintf("<p>%d of %d tests failed</p>\n", failures, total)
	printLinkToReport(suites.TestSuites)
	for _, s := range suites.TestSuites {
		if s.Failures > 0 || s.Errors > 0 {
			printSuiteHeader(s)
			for _, c := range s.TestCases {
				if c.Failure != nil || c.Error != nil {
					printTest(s, c)
				}
			}
		}
		printGatherLinks(s)
	}

	for _, s := range suites.TestSuites {
		printSuiteHeader(s)
		for _, c := range s.TestCases {
			if c.Failure == nil && c.Error == nil {
				printTest(s, c)
			}
		}
	}
	output += "</body>"
	output += "</html>"
	return output, nil
}

func printTest(testSuite reporters.JUnitTestSuite, testCase reporters.JUnitTestCase) {
	// regexp for replacing HTML tags in the log
	re := regexp.MustCompile(`<\/?[^>]+(>|$)`)
	id := fmt.Sprintf("%s.%s.%s", testSuite.Name, testCase.Classname, testCase.Name)
	class, text := "passed", "Pass"
	failure := testCase.Failure
	tcError := testCase.Error
	if tcError != nil {
		class, text = "error", "Error"
	}
	if failure != nil {
		class, text = "failed", "Fail"
	}
	skipped := testCase.Skipped
	if skipped != nil {
		class, text = "skipped", "Skip"
	}

	output += fmt.Sprintf("<div class='%s' id='%s'>\n", class, "div-"+id)

	output += fmt.Sprintf("<label for='%s' class='toggle'>%s<span class='badge'>%s</span></a></label>\n", id, testCase.Name, text)
	output += fmt.Sprintf("<input type='checkbox' name='one' id='%s' class='hide-input'>", id)
	output += "<div class='toggle-el'>\n"

	d := time.Duration(testCase.Time * float64(time.Second)).Round(time.Second)
	output += fmt.Sprintf("<p class='duration' title='Test duration'>Test duration: %v</p>\n", d)
	if tcError != nil {
		tcError.Message = re.ReplaceAllString(tcError.Message, "")
		tcError.Description = re.ReplaceAllString(tcError.Description, "")
		output += fmt.Sprintf("<div class='content'><b>Error message:</b> \n\n%s</div>\n", tcError.Message)
		output += fmt.Sprintf("<div class='content'><b>Error description:</b> \n\n%s</div>\n", tcError.Description)
	} else if failure != nil {
		failure.Message = re.ReplaceAllString(failure.Message, "")
		failure.Description = re.ReplaceAllString(failure.Description, "")
		output += fmt.Sprintf("<div class='content'><b>Failure message:</b> \n\n%s</div>\n", failure.Message)
		output += fmt.Sprintf("<div class='content'><b>Failure description:</b> \n\n%s</div>\n", failure.Description)
	} else if skipped != nil {
		skipped.Message = re.ReplaceAllString(skipped.Message, "")
		output += fmt.Sprintf("<div class='content'>%s</div>\n", skipped.Message)
	}
	if testCase.SystemErr != "" && testCase.Status != types.SpecStatePassed.String() {
		testCase.SystemErr = re.ReplaceAllString(testCase.SystemErr, "")
		output += fmt.Sprintf("<div class='content'><b>Log:</b> \n\n%s</div>\n", testCase.SystemErr)
	}

	output += "</div>\n"
	output += "</div>\n"
}

func printSuiteHeader(s reporters.JUnitTestSuite) {
	output += "<h4>"
	output += s.Name
	if len(s.Properties.Properties) != 0 {
		for _, p := range s.Properties.Properties {
			if strings.HasPrefix(p.Name, "coverage.") {
				v, _ := strconv.ParseFloat(p.Value, 32)
				output += fmt.Sprintf("<span class='coverage' title='%s'>%.0f%%</span>\n", p.Name, v)
			}
		}

	}
	output += "</h4>"
}

func printGatherLinks(s reporters.JUnitTestSuite) {
	if len(s.Properties.Properties) != 0 {
		for _, p := range s.Properties.Properties {
			if strings.Contains(p.Name, "gather") {
				output += fmt.Sprintf("<a href='%s'>Link to %s artifacts</a>\n", p.Value, p.Name)
			}
		}
	}
}

func printLinkToReport(suites []reporters.JUnitTestSuite) {
	for _, suite := range suites {
		if len(suite.Properties.Properties) != 0 {
			for _, p := range suite.Properties.Properties {
				if strings.Contains(p.Name, "html-report-link") {
					output += fmt.Sprintf("<a href='%s' target=”_blank” >Having trouble viewing this report? Click here to open it in another tab</a>\n", p.Value)
				}
			}
		}
	}
}
