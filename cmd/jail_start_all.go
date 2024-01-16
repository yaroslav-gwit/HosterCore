package cmd

// func _startAllJails(consoleLogOutput bool) error {
// 	jailList, err := GetAllJailsList()
// 	if err != nil {
// 		return err
// 	}

// 	startId := 0
// 	for _, v := range jailList {
// 		jailConfig, err := GetJailConfig(v, true)
// 		if err != nil {
// 			return err
// 		}
// 		jailOnline, err := checkJailOnline(jailConfig)
// 		if err != nil {
// 			return err
// 		}
// 		if jailOnline {
// 			continue
// 		}

// 		// Print out the output splitter
// 		if startId == 0 {
// 			_ = 0
// 		} else {
// 			fmt.Println("  ───────────")
// 		}

// 		if startId != 0 {
// 			time.Sleep(3 * time.Second)
// 		}

// 		err = _jailStart(v, consoleLogOutput)
// 		if err != nil {
// 			return err
// 		}

// 		startId += 1
// 	}

// 	return nil
// }
