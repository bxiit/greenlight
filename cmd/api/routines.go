package main

import (
	"github.com/bxiit/greenlight/internal/data"
	"time"
)

func (app *application) checkAndResendActivation() {
	for {
		users, err := app.models.UserInfos.FindNotActivatedAndExpired()
		if err != nil {
			app.logger.PrintFatal(err, nil)
			return
		}

		for _, user := range users {
			err := app.models.UserInfos.DeleteExpiredToken(user.ID)
			if err != nil {
				println("DeleteExpiredToken IS WRONG")
				return
			}

			token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
			if err != nil {
				return
			}

			app.background(func() {
				data := map[string]any{
					"activationToken": token.Plaintext,
					"userInfoID":      user.ID,
				}
				err := app.mailer.Send(user.Email, "user_welcome.tmpl", data)
				if err != nil {
					app.logger.PrintError(err, nil)
				}
			})
		}
		time.Sleep(time.Hour)
	}
}
