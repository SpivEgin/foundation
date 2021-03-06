package friendmail

import (
	"github.com/ottemo/foundation/env"
)

// setupConfig setups package configuration values for a system
func setupConfig() error {
	if config := env.GetConfig(); config != nil {
		err := config.RegisterItem(env.StructConfigItem{
			Path:        ConstConfigPathFriendMail,
			Value:       nil,
			Type:        env.ConstConfigTypeGroup,
			Editor:      "",
			Options:     nil,
			Label:       "Refer-A-Friend",
			Description: "Referal program",
			Image:       "",
		}, nil)

		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = config.RegisterItem(env.StructConfigItem{
			Path:        ConstConfigPathFriendMailEmailSubject,
			Value:       "Email friend",
			Type:        env.ConstConfigTypeVarchar,
			Editor:      "line_text",
			Options:     nil,
			Label:       "Email subject",
			Description: "Email subject for the friend form",
			Image:       "",
		}, nil)

		if err != nil {
			return env.ErrorDispatch(err)
		}

		err = config.RegisterItem(env.StructConfigItem{
			Path: ConstConfigPathFriendMailEmailTemplate,
			Value: `Dear {{.friend_name}}
<br />
<br />
Your friend sent you an email:
{{.content}}`,
			Type:        env.ConstConfigTypeHTML,
			Editor:      "multiline_text",
			Options:     nil,
			Label:       "Email Body",
			Description: "Email body template for the friend form",
			Image:       "",
		}, nil)

		if err != nil {
			return env.ErrorDispatch(err)
		}
	} else {
		err := env.ErrorNew(ConstErrorModule, env.ConstErrorLevelStartStop, "81e49a2f-906d-40ed-9a47-bb2b9c5e8f40", "Unable to obtain configuration for Friend Mail")
		return env.ErrorDispatch(err)
	}

	return nil
}
