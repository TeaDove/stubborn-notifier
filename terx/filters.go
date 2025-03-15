package terx

func FilterCommand(command string) func(c *Context) bool {
	return func(c *Context) bool {
		return c.Command == command
	}
}

func FilterNotCommand() func(c *Context) bool {
	return func(c *Context) bool {
		return c.Command == ""
	}
}

func FilterIsMessage() func(c *Context) bool {
	return func(c *Context) bool {
		return c.Update.Message != nil
	}
}

func FilterIsCallback() func(c *Context) bool {
	return func(c *Context) bool {
		return c.Update.CallbackQuery != nil
	}
}
