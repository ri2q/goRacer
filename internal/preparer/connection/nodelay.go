package preparer

func (c *ConnConfig) EnableDelay() error {
	if err := c.TcpConn.SetNoDelay(false); err != nil {
		return err
	}
	return nil
}

func (c *ConnConfig) DisableDelay() error {
	if err := c.TcpConn.SetNoDelay(true); err != nil {
		return err
	}
	return nil
}
