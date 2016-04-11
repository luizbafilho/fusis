package steps

import "fmt"

type step1 struct {
}

func (t step1) Do(prevResul Result) (Result, error) {
	return 10, nil
}

func (t step1) Undo() error {
	return nil
}

type step2 struct {
}

func (t step2) Do(prevResul Result) (Result, error) {
	return 20, nil
}

func (t step2) Undo() error {
	return nil
}

type errorStep struct {
}

func (t errorStep) Do(prevResul Result) (Result, error) {
	return nil, fmt.Errorf("Error")
}

func (t errorStep) Undo() error {
	return nil
}
