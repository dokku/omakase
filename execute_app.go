package main

type AppTask struct {
	context DokkuApp
}

func (t AppTask) Execute(context struct{}) error {
	return nil
}
