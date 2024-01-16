// Copyright 2023s The Hoster Authors. All rights reserved.
// Use of this source code is governed by a Apache License 2.0
// license that can be found in the LICENSE fil

package console_login

import (
	"time"

	"github.com/gcla/gowid"
)

type HomeController struct {
	view        *HomeView
	ticker      *time.Ticker
	lockTimeout int
}

func NewHomeController() *HomeController {
	res := &HomeController{nil, nil, sessionTimeDefault}
	view := NewHomeView(res)
	res.view = view

	return res
}

func (c *HomeController) ResetSessionTime(app gowid.IApp) {
	c.view.sessionTime = c.lockTimeout
	c.view.UpdateSessionTime(app)
}

func (c *HomeController) AnimateSessionTime(app gowid.IApp) {
	c.ticker = time.NewTicker(time.Second * 1)
	go func() {
		for range c.ticker.C {
			app.Run(gowid.RunFunction(func(app gowid.IApp) {
				c.view.UpdateSessionTime(app)
				app.Redraw()
			}))
		}
	}()
}

func (g *HomeController) StopAnimation() {
	g.ticker.Stop()
}

func (c *HomeController) GetSessionTime() int {
	sessionTime := host_config.ConsolePanel.SessionTime
	if sessionTime == 0 {
		return sessionTimeDefault
	}
	return sessionTime
}
