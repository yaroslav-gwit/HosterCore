package HosterLocations

func GetBinaryFolders() (r []string) {
	r = []string{
		"/opt/hoster-core",
		"/opt/hoster",
		"/usr/local/bin",
		"/bin",
		"/root/hoster",
	}

	return
}

func GetConfigFolders() (r []string) {
	r = []string{
		"/opt/hoster-core/config_files",
		"/opt/hoster/config_files",
		"/usr/local/hoster",
		"/etc/hoster",
		"/root/hoster/config_files",
	}

	return
}
