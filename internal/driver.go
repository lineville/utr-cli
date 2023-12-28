package internal

import (
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func CreateDriver() (selenium.WebDriver, error) {
	// TODO need to get appropriate chromedriver for the platform
	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{}
	caps.AddChrome(chrome.Capabilities{Args: []string{
		"window-size=1920x1080",
		"--no-sandbox",
		"--disable-dev-shm-usage",
		"disable-gpu",
		"--headless",
	}})

	driver, err := selenium.NewRemote(caps, "")
	if err != nil {
		panic(err)
	}
	return driver, nil
}
