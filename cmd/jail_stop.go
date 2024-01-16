package cmd

// func _jailStop(jailName string, logActions bool) error {
// 	jailConfig, err := GetJailConfig(jailName, false)
// 	if err != nil {
// 		return err
// 	}

// 	if logActions {
// 		emojlog.PrintLogMessage("Stopping the Jail: "+jailName, emojlog.Info)
// 	}

// 	_ = os.Remove("/etc/jail.conf")
// 	input, err := os.ReadFile(jailConfig.JailFolder + "jail_temp_runtime.conf")
// 	if err != nil {
// 		return err
// 	}
// 	err = os.WriteFile("/etc/jail.conf", input, 0644)
// 	if err != nil {
// 		return err
// 	}

// 	emojlog.PrintLogMessage(fmt.Sprintf("Stopping a Jail: %s. Please give it a moment...", jailName), emojlog.Debug)
// 	out, err := exec.Command("service", "jail", "onestop", jailName).CombinedOutput()
// 	if err != nil {
// 		errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
// 		return fmt.Errorf("%s", errorValue)
// 	}

// 	return nil
// }
