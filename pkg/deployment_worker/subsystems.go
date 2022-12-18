package deployment_worker

func getCreatedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		K8sCreated,
		NPMCreated,
		HarborCreated,
	}
}

func getDeletedConfirmers() []func(string) (bool, error) {
	return []func(string) (bool, error){
		K8sDeleted,
		NPMDeleted,
		HarborDeleted,
	}
}

func Created(name string) bool {
	confirmers := getCreatedConfirmers()
	for _, confirmer := range confirmers {
		created, _ := confirmer(name)
		if !created {
			return false
		}
	}
	return true
}

func Deleted(name string) bool {
	confirmers := getDeletedConfirmers()
	for _, confirmer := range confirmers {
		deleted, _ := confirmer(name)
		if !deleted {
			return false
		}
	}
	return true
}
