/*config read to verify normal user*/
package main

import "github.com/arstercz/goconfig"

func VerifyOk(username string, passinfo string) bool {
	if len(username) <= 0 || len(passinfo) <= 0 {
		return false
	}
	c, err := goconfig.ReadConfigFile(config.Conf)
	pass, err := c.GetString(username, "pass")
	if err != nil {
		return false
	}
	if pass == passinfo {
		return true
	}
	return false
}
