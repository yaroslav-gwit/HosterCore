package cmd

// Destroys any existing Jail dataset. Returns an error, if there was an issue running `zfs destroy`, or if the dataset doesn't exist
// func executeJailDestroy(jailName string, consoleLogOutput bool) error {
// 	jailConfig, err := GetJailConfig(jailName, false)
// 	if err != nil {
// 		return err
// 	}

// 	jailOnline, err := checkJailOnline(jailConfig)
// 	if err != nil {
// 		return err
// 	}
// 	if jailOnline {
// 		return fmt.Errorf("the Jail \"%s\" is online! Cannot remove online Jails", jailName)
// 	}

// 	// Get the parent dataset
// 	out, err := exec.Command("zfs", "get", "-H", "origin", jailConfig.ZfsDatasetPath).CombinedOutput()
// 	// Correct value:
// 	// NAME                           PROPERTY  VALUE                                                                           SOURCE
// 	//    [0]                             [1]           [2]                                                                     [3]
// 	// zroot/vm-encrypted/twelveFour	origin	zroot/vm-encrypted/jail-template-12.4-RELEASE@deployment_twelveFour_qc5q7u6khy	-
// 	// Empty value
// 	// zroot/vm-encrypted/wordpress-one	origin	         -	                                                                     -
// 	if err != nil {
// 		errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
// 		return fmt.Errorf("%s", errorValue)
// 	}

// 	reSpaceSplit := regexp.MustCompile(`\s+`)
// 	parentDataset := reSpaceSplit.Split(strings.TrimSpace(string(out)), -1)[2]
// 	// EOF Get the parent dataset

// 	// Remove the Jail dataset
// 	out, err = exec.Command("zfs", "destroy", jailConfig.ZfsDatasetPath).CombinedOutput()
// 	if err != nil {
// 		errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
// 		return fmt.Errorf("%s", errorValue)
// 	}
// 	if consoleLogOutput {
// 		emojlog.PrintLogMessage(fmt.Sprintf("Jail dataset has been destroyed: %s", jailConfig.ZfsDatasetPath), emojlog.Changed)
// 	}
// 	// EOF Remove the Jail dataset

// 	// Remove the parent dataset if it exists
// 	if len(parentDataset) > 1 {
// 		out, err := exec.Command("zfs", "destroy", parentDataset).CombinedOutput()
// 		if err != nil {
// 			errorValue := "FATAL: " + strings.TrimSpace(string(out)) + "; " + err.Error()
// 			return fmt.Errorf("%s", errorValue)
// 		}
// 		if consoleLogOutput {
// 			emojlog.PrintLogMessage(fmt.Sprintf("Jail dataset parent snapshot has been destroyed: %s", parentDataset), emojlog.Changed)
// 		}
// 	}
// 	// EOF Remove the parent dataset if it exists

// 	if consoleLogOutput {
// 		emojlog.PrintLogMessage(fmt.Sprintf("Jail has been removed: %s", jailName), emojlog.Info)
// 	}

// 	return nil
// }
