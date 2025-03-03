package service

import "fmt"

func (c *GithubConnectorService) UpdateGithubConnectorRequest(InstallationID string,UserID string) error {
	connector, err := c.storage.GetAllConnectors(UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if len(connector) == 0 {
		fmt.Println("no connector found")
		return nil
	}

	err = c.storage.UpdateConnector(connector[0].ID.String(), InstallationID)
	if err != nil {
		return err
	}
	return nil
}
