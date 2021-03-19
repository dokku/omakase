package main

type SyncTask struct {
	context DokkuSync
}

func (t SyncTask) Execute(context struct{}) error {
	return nil
}
