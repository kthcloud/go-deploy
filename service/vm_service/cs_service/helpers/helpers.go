package helpers

func (client *Client) GetFreePort() (int, error) {
	return client.SsClient.GetFreePort(client.Zone.PortRange.Start, client.Zone.PortRange.End)
}
