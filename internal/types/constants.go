package types

import (
	"time"
)

var (
	// AppVersion Inject by -ldflags
	AppVersion   = "v5.0.0"
	BuildTags    = ""
	LastfmKey    = ""
	LastfmSecret = ""
)

const AppName = "musicfox"
const GroupID = "com.anhoder.musicfox"
const AppDescription = "<cyan>musicfox-deck - 磁带卡座风格的命令行网易云音乐</>"

// AppRepo is the GitHub repository of musicfox-deck, a fork of
// go-musicfox/go-musicfox: releases, update checks and bug reports go there.
const AppRepo = "Egdon/musicfox-deck"
const AppGithubUrl = "https://github.com/" + AppRepo
const AppLatestReleases = AppGithubUrl + "/releases/latest"
const AppCheckUpdateUrl = "https://api.github.com/repos/" + AppRepo + "/releases/latest"
const AppIssueUrl = AppGithubUrl + "/issues/new?template=1_bug_report.yml"
const LastfmAuthUrl = "https://www.last.fm/api/auth/?api_key=%s&token=%s"
const ProgressFullChar = "#"
const ProgressEmptyChar = "."
const StartupLoadingSeconds = 2
const StartupTickDuration = time.Millisecond * 16

const AppLocalDataDir = "go-musicfox"
const AppDBName = "musicfox"
const AppTomlFile = "config.toml"
const AppPrimaryRandom = "random"
const AppPrimaryColor = "#f90022"
const SubmitText = "确认"
const SearchPlaceholder = "搜索"
const SearchResult = "搜索结果"
const AppHttpTimeout = time.Second * 5

const MainLoadingText = "[加载中...]"
const MainPProfPort = 9876
const DefaultNotifyIcon = "logo.png"
const UNMDefaultSources = "kuwo"

const BeepPlayer = "beep"          // beep
const MpdPlayer = "mpd"            // mpd
const MpvPlayer = "mpv"            // mpv player
const OsxPlayer = "osx"            // osx
const WinMediaPlayer = "win_media" // win media player
const DlnaPlayer = "dlna"          // dlna

const BeepGoMp3Decoder = "go-mp3"
const BeepMiniMp3Decoder = "minimp3"

const MaxPlayErrCount = 3

const SearchPageSize = 100

const AppHelpTemplate = `%s

{{.Description}} (Version: <info>{{.Version}}</>)

<comment>Usage:</>
  {$binName} [Global Options...] <info>{command}</> [--option ...] [argument ...]

<comment>Global Options:</>
{{.GOpts}}
<comment>Available Commands:</>{{range $module, $cs := .Cs}}{{if $module}}
<comment> {{ $module }}</>{{end}}{{ range $cs }}
  <info>{{.Name | paddingName }}</> {{.UseFor}}{{if .Aliases}} (alias: <cyan>{{ join .Aliases ","}}</>){{end}}{{end}}{{end}}

  <info>{{ paddingName "help" }}</> Display help information

Use "<cyan>{$binName} {COMMAND} -h</>" for more information about a command
`
