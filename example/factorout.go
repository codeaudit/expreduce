package main

import (
	"fmt"
	"bytes"
	"flag"
	"strings"
	"regexp"
	"log"
	"github.com/corywalker/expreduce/expreduce"
)

var modules = flag.String("modules", "",
	"A regexp of modules to include, otherwise include all modules.")

func main() {
	flag.Parse()
	//es := expreduce.NewEvalState()
	var ModEx = regexp.MustCompile(*modules)
	for _, defSet := range expreduce.GetAllDefinitions() {
		if !ModEx.MatchString(defSet.Name) {
			continue
		}
		var b bytes.Buffer
		for _, def := range defSet.Defs {
			if def.Bootstrap {
				continue
			}
			if def.Usage != "" {
				b.WriteString(fmt.Sprintf("%s::usage = \"%v\";\n", def.Name, def.Usage))
				for _, r := range def.Rules {
					b.WriteString(fmt.Sprintf("%v := %v;\n", r.Lhs, r.Rhs))
				}
				b.WriteString(fmt.Sprintf("Attributes[%v] = {", def.Name))
				for _, a := range def.Attributes {
					b.WriteString(fmt.Sprintf("%s, ", a))
				}
				b.WriteString(fmt.Sprintf("Protected};\n"))
				var tests bytes.Buffer
				hasTests := false
				tests.WriteString(fmt.Sprintf("Tests`%v = {\n\t", def.Name))
				testCols := [][]expreduce.TestInstruction{
					def.SimpleExamples,
					def.FurtherExamples,
					def.Tests,
					def.KnownFailures,
					def.KnownDangerous,
				}
				testColNames := []string{
					"ESimpleExamples",
					"EFurtherExamples",
					"ETests",
					"EKnownFailures",
					"EKnownDangerous",
				}
				for i, testCol := range testCols {
					if len(testCol) > 0 {
						//fmt.Println(testCol, testColNames[i])
						if hasTests {
							tests.WriteString(fmt.Sprintf(", %v[\n", testColNames[i]))
						} else {
							tests.WriteString(fmt.Sprintf("%v[\n", testColNames[i]))
						}
						for ti, t := range testCol {
							tests.WriteString(fmt.Sprintf("\t\t"))
							if tSame, tIsSame := t.(*expreduce.SameTest); tIsSame {
								tests.WriteString(fmt.Sprintf("ESameTest[%v, %v]", tSame.Out, tSame.In))
							} else if tComment, tIsComment := t.(*expreduce.TestComment); tIsComment {
								tests.WriteString(fmt.Sprintf("EComment[\"%v\"]", tComment.Comment))
							} else if tString, tIsString := t.(*expreduce.StringTest); tIsString {
								tests.WriteString(fmt.Sprintf("EStringTest[\"%v\", \"%v\"]", tString.Out, tString.In))
							} else if tDiff, tIsDiff := t.(*expreduce.DiffTest); tIsDiff {
								tests.WriteString(fmt.Sprintf("EDiffTest[%v, %v]", tDiff.Out, tDiff.In))
							} else if tExampleOnly, tIsExampleOnly := t.(*expreduce.ExampleOnlyInstruction); tIsExampleOnly {
								tests.WriteString(fmt.Sprintf("EExampleOnlyInstruction[\"%v\", \"%v\"]", tExampleOnly.Out, tExampleOnly.In))
							} else if _, tIsResetState := t.(*expreduce.ResetState); tIsResetState {
								tests.WriteString(fmt.Sprintf("EResetState[]"))
							} else {
								tests.WriteString(fmt.Sprintf("%v", t))
								log.Fatalf("%v %v %v", t, defSet.Name, def.Name)
							}
							if ti != len(testCol)-1 {
								tests.WriteString(fmt.Sprintf(","))
							}
							tests.WriteString(fmt.Sprintf("\n"))
						}
						tests.WriteString(fmt.Sprintf("\t]"))
						hasTests = true
					}
				}
				tests.WriteString(fmt.Sprintf("\n};"))
				if hasTests {
					b.WriteString(fmt.Sprintf("%v\n", tests.String()))
				}
				b.WriteString(fmt.Sprintf("\n"))
			}
		}
		fmt.Printf("%s\n", strings.Replace(b.String(), "\t", "    ", -1))
	}
}
