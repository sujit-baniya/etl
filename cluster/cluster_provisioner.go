package cluster

func NewProvisioner() *Provisioner {
	provisioner := new(Provisioner)

	provisioner.RegisteredFunctions = make(map[string]*Cluster)
	provisioner.OperationalFunctions = make(map[string]*Cluster)
	provisioner.Configs = make(map[string]Config)
	provisioner.Registries = make(map[string]*Registry)

	return provisioner
}

func (provisioner *Provisioner) Register(function string, cluster Cluster, config ...Config) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.RegisteredFunctions[function]; found {
		return false
	}

	provisioner.RegisteredFunctions[function] = &cluster
	if len(config) > 0 {
		provisioner.Configs[function] = config[0]
	}

	provisioner.Registries[function] = NewRegistry()

	return true
}

func (provisioner *Provisioner) Function(identifier string) (*Cluster, *Config, *Registry, bool) {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	if _, found := provisioner.OperationalFunctions[identifier]; !found {
		return nil, nil, nil, false
	}

	cluster := provisioner.OperationalFunctions[identifier]
	config := provisioner.Configs[identifier]
	registry := provisioner.Registries[identifier]

	return cluster, &config, registry, true
}

func (provisioner *Provisioner) Mount(identifier string) bool {
	provisioner.mutex.Lock()
	defer provisioner.mutex.Unlock()

	// if the function has already been mounted and is in an operational state, we don't need to do anything
	if _, found := provisioner.OperationalFunctions[identifier]; found {
		return true
	}

	// if the function is not mounted, but exists as a registered function, mount it so that it can be provisioned
	if cluster, found := provisioner.RegisteredFunctions[identifier]; found {
		provisioner.OperationalFunctions[identifier] = cluster
	} else {
		return false
	}

	return true
}

func (provisioner *Provisioner) UnMount(identifier string) bool {

	if _, found := provisioner.OperationalFunctions[identifier]; !found {
		return false
	}

	delete(provisioner.OperationalFunctions, identifier)

	return true
}

func (provisioner Provisioner) IsMounted(identifier string) bool {
	_, found := provisioner.OperationalFunctions[identifier]
	return found
}
