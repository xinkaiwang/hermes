package common

var (
	Version   = "unkonwn" // 0.1.0
	GitCommit = "unkonwn" // 3cf1cec
	BuildTime = "unkonwn" // 2025-09-15_05:37:33
)

func GetVersion() string {
	return Version
}

func GetGitCommit() string {
	return GitCommit
}

func GetBuildTime() string {
	return BuildTime
}
