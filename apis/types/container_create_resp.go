package types

//ContainerCreateResp stands for response object of request /containers/create
type ContainerCreateResp struct {
	ID       string
	Name     string
	Warnings []string
}
