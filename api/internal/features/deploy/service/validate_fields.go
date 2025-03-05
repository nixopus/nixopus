package service

func (s *DeployService) IsNameAlreadyTaken(name string) (bool, error) {
	return s.storage.IsNameAlreadyTaken(name)
}

func (s *DeployService) IsDomainValid(domain string) (bool, error) {
	return s.storage.IsDomainValid(domain)
}

func (s *DeployService) IsDomainAlreadyTaken(domain string) (bool, error) {
	return s.storage.IsDomainAlreadyTaken(domain)
}

func (s *DeployService) IsPortAlreadyTaken(port int) (bool, error) {
	return s.storage.IsPortAlreadyTaken(port)
}
