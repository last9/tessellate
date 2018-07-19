package dispatcher


type Dispatcher interface {
	Dispatch(jobID, workspaceID string) error
}