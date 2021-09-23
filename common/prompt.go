package common

import (
	"errors"
	"github.com/manifoldco/promptui"
	"strings"
)

func ValidateEmptyInput(input string) error {
	if len(strings.TrimSpace(input)) < 1 {
		return errors.New("this input must not be empty")
	}
	return nil
}

func PromptString(name string, currentValue string, isConfirm bool) (string, error) {
	prompt := promptui.Prompt{
		Label:    name,
		Validate: ValidateEmptyInput,
		Default: currentValue,
		IsConfirm: isConfirm,
	}

	return prompt.Run()
}

func PromptSelect(name string, items []string) (int, string, error){
	prompt := promptui.Select{
		Label: name,
		Items: items,
	}

	return prompt.Run()
}

//func PromptInteger(name string) (int64, error) {
//	prompt := promptui.Prompt{
//		Label:    name,
//		Validate: ValidateIntegerNumberInput,
//	}
//
//	promptResult, err := prompt.Run()
//	if err != nil {
//		return 0, err
//	}
//
//	parseInt, _ := strconv.ParseInt(promptResult, 0, 64)
//	return parseInt, nil
//}