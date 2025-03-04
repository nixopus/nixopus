package service

import "fmt"

// UpdateGithubConnectorRequest updates the GitHub connector request for the given user ID.
//
// The method first retrieves all GitHub connectors associated with the user ID.
// If no connectors are found, the method simply returns.
//
// Otherwise, the method takes the ID of the first connector and updates the
// associated GitHub app ID with the provided InstallationID.
//
// If any errors occur during the update process, the method returns the error.
func (c *GithubConnectorService) UpdateGithubConnectorRequest(InstallationID string, UserID string) error {
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
