package utils

type MyAnts interface {
	UseMyAnts() error
}

type MyAntsImpl struct {
}

func (MyAntsImpl) UseMyAnts() error {

	return nil
}
