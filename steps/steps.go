package steps

// Result is the value returned by Forward. It is used in the call of the next
// action, and also when rolling back the actions.
type Result interface{}

type Step interface {
	Do(prevResult Result) (Result, error)
	Undo() error
}

type Sequence struct {
	Steps []Step
}

func NewSequence(steps ...Step) *Sequence {
	return &Sequence{
		Steps: steps,
	}
}

func (s *Sequence) Execute() error {
	var prevResult Result
	prevResult, err := s.Steps[0].Do(nil)
	if err != nil {
		return err
	}

	for k, step := range s.Steps[1:] {
		prevResult, err = step.Do(prevResult)
		if err != nil {
			s.rollback(k)
			return err
		}
	}

	return nil
}

func (s *Sequence) rollback(point int) {
	for i := point; i >= 0; i-- {
		s.Steps[i].Undo()
	}
}

func (s *Sequence) RollbackAll() error {
	for i := len(s.Steps); i >= 0; i-- {
		if err := s.Steps[i].Undo(); err != nil {
			return err
		}
	}
	return nil
}
