package service

import (
	"errors"
	"fmt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"os/exec"
	"regexp"
	"strings"
)

type ShellResponse struct {
	Stdout string
	Stderr string
}

func injectArguments(cmd string, inputs map[string]interface{}) (string, error) {
	re := regexp.MustCompile("\\{\\{([a-zA-Z0-9]+)}}")
	erroredArgs := make([]string, 0)
	newStr := re.ReplaceAllStringFunc(cmd, func(match string) string {
		key := re.FindStringSubmatch(match)[1]
		if val, exists := inputs[key]; exists {
			return fmt.Sprintf("%v", val)
		}
		erroredArgs = append(erroredArgs, match)
		return match
	})
	if len(erroredArgs) > 0 {
		return "", fmt.Errorf("following args were not found: %s", strings.Join(erroredArgs, ", "))
	}
	return newStr, nil
}

/*
Eval
Runs an evaluation on an expression using bc command

	inputs: {
		exp: "1 + {{a}}",
		a: "223"
	}

Responds "224", nil
*/
func Eval(inputs map[string]interface{}, span trace.Span) (string, error) {
	exp, exists := inputs["exp"]
	if !exists {
		span.RecordError(fmt.Errorf("exp field not found in inputs: %v", inputs))
		return "", errors.New("exp field is required")
	}
	expression, err := injectArguments(exp.(string), inputs)
	if err != nil {
		span.RecordError(err)
		return "", err
	}
	span.SetAttributes(attribute.String("exp", expression))
	out, err := exec.Command("bash", "-c", fmt.Sprintf("echo \"%s\" | bc", expression)).Output()
	if err != nil {
		span.RecordError(err)
		return "", err
	}
	return strings.TrimRight(string(out), "\n"), nil
}

/*
RunShell
Runs bash

		 inputs: {
			cmd: "echo {{1}} {{2}}",
			1: "aaaaa",
			2: "bbbbb"
		 }
		 span
		 For normal usage, the error command is only reserved for mid execution mistakes
	     Any output from Stderr is not returned by error, instead from ShellResponse
*/
func RunShell(inputs map[string]interface{}, span trace.Span) (ShellResponse, error) {
	cmd, exists := inputs["cmd"]
	if !exists {
		return ShellResponse{}, fmt.Errorf("the cmd field is required")
	}
	finalCmd, err := injectArguments(cmd.(string), inputs)
	if err != nil {
		span.RecordError(err)
		return ShellResponse{}, err
	}
	span.SetAttributes(attribute.String("cmd", finalCmd))
	out, err := exec.Command("bash", "-c", finalCmd).Output()
	if err != nil {
		span.RecordError(err)
		fmt.Printf("Error executing command %s", err)
		var exitError *exec.ExitError
		switch {
		case errors.As(err, &exitError):
			if err.(*exec.ExitError).ExitCode() == 127 {
				return ShellResponse{
					Stderr: "Unknown command",
				}, nil
			}
			return ShellResponse{
				Stderr: strings.TrimRight(string(err.(*exec.ExitError).Stderr), "\n"),
			}, nil
		default:
			return ShellResponse{}, fmt.Errorf("error executing command %s", err)
		}
	}
	return ShellResponse{
		Stdout: strings.TrimRight(string(out), "\n"),
	}, nil
}
