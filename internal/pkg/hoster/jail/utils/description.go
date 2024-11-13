package HosterJailUtils

import "errors"

func UpdateDescription(jailName string, description string) error {
	jail, err := InfoJsonApi(jailName)
	if err != nil {
		return err
	}

	jailConfLoc := jail.Simple.Mountpoint + "/" + jail.Name + "/" + JAIL_CONFIG_NAME
	if len(description) < 1 {
		return errors.New("description is empty")
	}
	if len(description) > 255 {
		return errors.New("description is too long")
	}

	jail.JailConfig.Description = description
	err = ConfigFileWriter(jail.JailConfig, jailConfLoc)
	if err != nil {
		return err
	}

	return nil
}
