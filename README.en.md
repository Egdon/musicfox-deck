<h1 align="center">tapefox</h1>

<p align="center">
  <a href="https://github.com/Egdon/tapefox/releases"><img alt="Release" src="https://img.shields.io/github/v/release/Egdon/tapefox?include_prereleases&label=release"></a>
  <img alt="macOS" src="https://img.shields.io/badge/macOS-Apple%20silicon%20%2F%20Intel-000000">
  <img alt="Windows" src="https://img.shields.io/badge/Windows-10%20%2F%2011-0078d4">
  <img alt="Linux" src="https://img.shields.io/badge/Linux-x86__64%20%2F%20ARM64-fcc624">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26-00add8">
  <a href="https://github.com/go-musicfox/go-musicfox"><img alt="Based on go-musicfox" src="https://img.shields.io/badge/based%20on-go--musicfox%20v5.1.0-d2714f"></a>
</p>

<p align="center">A terminal client for NetEase Cloud Music with a cassette-deck interface: a 3D cassette rendered in real time sits at its center, and page changes, menus and popups all have transitions.</p>

<p align="center">
  <a href="https://github.com/Egdon/tapefox/releases/latest"><strong>Download the latest release</strong></a>
  ·
  <a href="./CHANGELOG.md">Changelog</a>
  ·
  <a href="https://github.com/Egdon/tapefox/issues">Issues</a>
  ·
  <a href="./README.md">简体中文</a>
</p>

## Preview

### Main screen: tracklist and 3D cassette panel

![tapefox main screen: tracklist and 3D cassette panel](./docs/assets/deck-panel.gif)

### Now playing: the floating cassette and word-by-word lyrics

![tapefox now-playing page: floating cassette and word-by-word lyrics](./docs/assets/now-playing.gif)

<table>
  <tr>
    <td width="50%" align="center"><strong>Entering the page: the cassette grows onto the stage</strong></td>
    <td width="50%" align="center"><strong>Track change: the cassette flips over</strong></td>
  </tr>
  <tr>
    <td><img src="./docs/assets/morph.gif" alt="The cassette grows from the panel onto the stage" width="420"></td>
    <td><img src="./docs/assets/flip.gif" alt="The cassette flips over on a track change" width="420"></td>
  </tr>
  <tr>
    <td width="50%" align="center"><strong>Startup: the cassette turns in from the distance</strong></td>
    <td width="50%" align="center"><strong>Menus: rows slide in</strong></td>
  </tr>
  <tr>
    <td><img src="./docs/assets/startup.gif" alt="The deck startup animation" width="420"></td>
    <td><img src="./docs/assets/menus.gif" alt="Menu rows slide in" width="420"></td>
  </tr>
  <tr>
    <td colspan="2" align="center"><strong>Narrow window: player strip and poster layout</strong></td>
  </tr>
  <tr>
    <td colspan="2" align="center"><img src="./docs/assets/narrow.gif" alt="Player strip and poster layout in a narrow window" width="660"></td>
  </tr>
</table>

These recordings come from the program running for real. The cassette is drawn with octant characters, as in Ghostty, kitty and similar terminals; other terminals fall back to coarser characters, see [Terminals and display](#terminals-and-display). The interface is in Simplified Chinese.

## About

tapefox is a fork of [go-musicfox](https://github.com/go-musicfox/go-musicfox) v5.1.0 that reworks the main screen, the now-playing page and the motion: it adds a 3D cassette renderer, a tracklist with a cassette panel, a stage-style now-playing page, transitions throughout and the Warm Night theme, and ships prebuilt packages for macOS, Windows and Linux. Everything else comes from go-musicfox, and the two share their configuration and data directory.

This repository is not the official go-musicfox repository, and its original authors do not support or endorse this project. All changes are listed in the [changelog](./CHANGELOG.md).

## Downloads and platforms

| Platform | Installation | Architectures |
| --- | --- | --- |
| macOS 12 or later | Homebrew: `brew install egdon/tap/tapefox` | Apple silicon, Intel |
| Windows 10 / 11 | Scoop: `scoop bucket add egdon https://github.com/Egdon/scoop-bucket`, then `scoop install tapefox` | x64, ARM64 |
| Debian / Ubuntu | Download the `.deb`, then `sudo apt install ./tapefox_*_amd64.deb` | x86_64, ARM64 |
| Fedora / openSUSE | Download the `.rpm`, then `sudo dnf install ./tapefox-*.x86_64.rpm` | x86_64, ARM64 |
| Arch Linux | Download the `.pkg.tar.zst`, then `sudo pacman -U tapefox-*-x86_64.pkg.tar.zst` | x86_64, ARM64 |
| Other Linux distributions | Download `linux_amd64.tar.gz` and put `tapefox` on your `PATH` | x86_64, ARM64 |

Every package and archive is published on [GitHub Releases](https://github.com/Egdon/tapefox/releases/latest) together with a `SHA256SUMS` file. `amd64` and `x86_64` in file names both mean x86_64; `arm64` and `aarch64` both mean ARM64. Once installed, run `tapefox` in a terminal.

### macOS

Homebrew is the recommended method:

```sh
brew install egdon/tap/tapefox
```

To install manually, download `darwin_arm64.tar.gz` (Apple silicon) or `darwin_amd64.tar.gz` (Intel). The binary is not notarized by Apple: a file downloaded with a browser needs its quarantine attribute removed before the first run, or permission in System Settings → Privacy & Security:

```sh
xattr -d com.apple.quarantine ./tapefox
```

Binaries installed by Homebrew carry no quarantine attribute and need no such step. `darwin_*_macapp.zip` holds the .app version, whose notifications show the app icon.

### Windows

Scoop is the recommended method; `scoop update tapefox` upgrades it later:

```powershell
scoop bucket add egdon https://github.com/Egdon/scoop-bucket
scoop install tapefox
```

To install manually, download `windows_amd64.zip` (x64) or `windows_arm64.zip` (ARM64), extract it and run `tapefox.exe` in [Windows Terminal](https://aka.ms/terminal), or add its directory to your `PATH`.

The binary is not signed. Started from File Explorer, it may bring up a SmartScreen prompt, where "More info" → "Run anyway" lets it run; with Smart App Control turned on in Windows 11, unsigned programs cannot run. If security software reports it, check the file against `SHA256SUMS` on the Releases page.

### Linux

The packages and the archives contain the same binary, which needs:

- glibc 2.34 or later (Ubuntu 22.04, Debian 12, Fedora 35, RHEL 9 and later releases)
- ALSA (`libasound.so.2`)

The deb, rpm and Arch packages declare these dependencies.

Web login needs WebKitGTK 4.1 (Debian / Ubuntu: `libwebkit2gtk-4.1-0`; Fedora: `webkit2gtk4.1`; Arch Linux: `webkit2gtk-4.1`). The deb and rpm packages recommend it; without it, the login page offers QR code login only.

The beep engine needs `libasound2-plugins` under WSL (see [microsoft/wslg#864](https://github.com/microsoft/wslg/issues/864)).

## Highlights

- 3D cassette: rendered in real time by perspective ray casting, at 2×4 subpixels per cell in terminals with octant characters. It floats while playing, tilts toward the mouse, flips over on a track change and comes to rest when paused; the tape on its reels follows the playback position.
- Tracklist and cassette panel: the tracklist fills the height, and the panel on the right shows the current track, word-by-word lyrics and the play queue. In windows narrower than 110 columns, the panel becomes a player strip at the bottom.
- Stage-style now-playing page: a stage layout in wide windows, a poster layout in narrow ones and a landscape layout in short ones; the cassette morphs between the panel and the stage when entering and leaving.
- Transitions: page changes, menus, popups, notifications, the selection and the startup screen have 150–250 ms transitions without overshoot. Motion can be reduced to fades or turned off.
- Word-by-word lyrics: NetEase YRC word timing, lighting up each word as it is sung.
- Themes: the default Warm Night pairs a warm grey background with beige text and keeps signal orange for what is in progress; Default, Dracula, Nord, Gruvbox and others are built in, and custom themes are supported.
- Background loading: track changes and playlists load in the background, so the interface never waits on the network.
- Audio quality: from standard up to lossless, Hi-Res and master quality, with UnblockNeteaseMusic support.
- Playback engines: beep, mpv, MPD, the native macOS and Windows engines, and DLNA casting.
- System integration: MPRIS (Linux), Now Playing and desktop lyrics (macOS), system media controls (Windows), global hotkeys and Last.fm scrobbling.
- Mouse: click, double-click, scroll wheel, hover and right-click menus.

## Terminals and display

The cassette is drawn with Unicode block characters, so its detail depends on the characters a terminal supports. With `[main] blockGlyphs = "auto"` (the default) the characters follow the terminal:

| Terminal | Characters | Subpixels per cell |
| --- | --- | --- |
| Ghostty, kitty, foot, and VTE terminals such as recent GNOME Terminal | Octants | 2×4 |
| WezTerm, iTerm2, Alacritty, Konsole, Windows Terminal | Sextants | 2×3 |
| Inside tmux, and other terminals | Quadrants | 2×2 |

A monospaced font is required. The macOS Terminal app and the Windows Command Prompt (CMD) are not supported. Over SSH, `[main] motion = "subtle"` is recommended.

## Data storage

The configuration file is `config.toml` in the user configuration directory; the [default configuration](./utils/filex/embed/config.toml) lists every setting, with comments in Chinese.

| System | User configuration directory |
| --- | --- |
| macOS | `$HOME/Library/Application Support/go-musicfox` |
| Linux | `$XDG_CONFIG_HOME/go-musicfox` or `$HOME/.config/go-musicfox` |
| Windows | `%AppData%\go-musicfox` |

The `MUSICFOX_ROOT` environment variable selects another directory. Every installation method, Scoop included, uses these directories.

### Migrating from go-musicfox

tapefox shares the configuration and data directory with go-musicfox, so the login, configuration and cache stay in use. Their commands are `tapefox` and `tapefox`, so both can be installed side by side; when both run at once, the one started second works on a temporary copy of the database and its changes are not kept.

go-musicfox installed with Scoop keeps its data in `~\scoop\persist\go-musicfox\data`; copy the contents of that directory to `%AppData%\go-musicfox` to keep using them in tapefox.

Settings that are new in tapefox can be added to an existing configuration with `tapefox upgrade-config`; see [Configuration upgrade](#configuration-upgrade).

## Usage

```sh
tapefox
```

<details>
<summary>

### Notes
</summary>

- A monospaced font is required, or the two-column layout may be misaligned; setting `doubleColumn` to `false` also avoids it.
- If the cursor moves, tracks change or playback pauses unexpectedly, set `enableMouseEvent` to `false`.
- The macOS Terminal app and the Windows Command Prompt (CMD) are not supported; see [go-musicfox#99](https://github.com/go-musicfox/go-musicfox/issues/99). Recommended terminals: Ghostty, iTerm2 or kitty on macOS; Ghostty, kitty or foot on Linux; Windows Terminal on Windows.
- Files follow the XDG Base Directory specification, so some paths differ from older versions; see [XDG support](./docs/xdg_support.md) (in Chinese).
- The configuration file uses [TOML](https://toml.io/).

</details>
<details>
<summary>

### DLNA casting
</summary>

The DLNA engine casts music to DLNA devices such as smart TVs, speakers and game consoles.

1. **Find the device URL**: install `gupnp-tools` (macOS: `brew install gupnp-tools`; Debian / Ubuntu: `sudo apt install gupnp-tools`), run `gupnp-universal-cp` to scan for devices, and copy the device's description URL (such as `http://192.168.1.100:49152/description.xml`).
2. **Configure** in `config.toml`:
   ```toml
   [player]
   engine = "dlna"

   [player.dlna]
   deviceUrl = "http://<device IP>:<port>/description.xml"
   localIP = "<this computer's LAN IP>"  # for example 192.168.1.50
   ```
3. **Play**: songs chosen after startup play on the device.

</details>
<details>
<summary>

### Themes
</summary>

Themes can be switched at runtime and customized, and several are built in.

#### Built-in themes

| Theme | Description |
|---|---|
| Warm Night (default) | Warm grey background and beige text; signal orange marks what is in progress, grey green marks options that are on. Shares its palette with the now-playing page; designed for use at night |
| Default | Classic light theme |
| Transparent | Transparent background that keeps the terminal's |
| Dark Transparent | Dark transparent theme |
| Dracula | Dracula colors |
| Nord | Nord's cool colors |
| Gruvbox | Gruvbox's retro warm colors |

#### Usage

1. **Theme at startup**:
   ```toml
   [theme]
   activeTheme = "Dracula"
   ```
2. **Switch at runtime**: choose 切换主题 (switch theme) in the right-click menu, or bind a key to `switchTheme` in `[keybindings.app]` (for example `switchTheme = "ctrl+t"`).
3. **Custom themes**: create a `themes/` directory in the user configuration directory, with one `.toml` file per theme, for example:
   - macOS: `~/Library/Application Support/go-musicfox/themes/mytheme.toml`
   - Linux: `~/.config/go-musicfox/themes/mytheme.toml`
   - Windows: `%AppData%\go-musicfox\themes\mytheme.toml`

   A theme file needs a `name` field and a `[dark]` or `[light]` table; see the [built-in themes](./utils/filex/embed/themes/) for the format.

</details>
<details>
<summary>

### Cassette deck interface
</summary>

Enabled by `[theme] layout = "deck"` (the default); `"classic"` restores the go-musicfox layout.

- **Tracklist**: four columns (number, title, artist, duration) under a section header such as `07 ── 飙升榜  [更新27首] ……… 100 首`. The selection marker slides as the selection moves.
- **Cassette panel**: shown on the right in windows of at least 110 columns × 24 rows, with a small cassette, the title, word-by-word lyrics and the play queue. Clicking the cassette opens the now-playing page, clicking a track in the queue plays it, and the scroll wheel sets the volume.
- **Player strip**: in narrower or shorter windows, the panel becomes a strip below the tracklist with a small cassette, the title and artist and the current lyric line, and the position, play mode and volume on the right.

Related settings:

```toml
[main]
motion = "full"        # motion: full (default) / subtle (fades only) / off
blockGlyphs = "auto"   # block characters for the cassette: auto / octant / sextant / quadrant / half

[theme]
layout = "deck"        # deck / classic
activeTheme = "Warm Night"

[startup]
animation = "deck"     # the cassette turns in from the distance and the wordmark appears letter by letter
```

See [Terminals and display](#terminals-and-display) for how `blockGlyphs = "auto"` chooses.

</details>
<details>
<summary>

### Now-playing page
</summary>

Press `i` or click the cassette on the main screen to open it; `i` or `Esc` goes back.

- **Wide windows** (at least 100 columns × 28 rows): the stage on the left shows the cassette scaled to the window; the numbered column on the right lists the title, artist, album and quality under `02 TRACK` and word-by-word lyrics under `03 LYRICS`; a seven-segment counter and `04 UP NEXT` sit below.
- **Narrow windows**: a poster layout with the cassette centered and the title, artist, current lyrics, time and progress bar below it.
- **Short windows**: the cassette on the left; the track, time and lyrics on the right.
- **Cassette**: floats slowly while playing, tilts toward the mouse, flips over on a track change and comes to rest when paused. Its label carries the A/B side and ruled lines.
- **Word-by-word lyrics**: lyrics with word timing (YRC) light up word by word; lyrics with line timing only are highlighted by line.
- **Status**: indicator lights in the header show shuffle, repeat, single, liked, lyrics and high quality; a VCR-style OSD shows the playback state and volume; `L` toggles full-screen lyrics; a NO SIGNAL screen appears when playback stops with an empty queue.
- **Keys on the page**: `←` / `→` seek 5 seconds back / forward, `↑` / `↓` set the volume; other keys keep their bindings and act on the playing song.
- **Mouse**: click the progress bar to seek, the key bar at the bottom to control playback, or a track in the queue to play it; the scroll wheel sets the volume.

Changing the volume, pausing, seeking or switching the play mode on the main screen shows the same style of OSD in the top right corner (off with `[main.notification] inApp = false`).

</details>
<details>
<summary>

### Live spectrum
</summary>

Shows a live audio spectrum in the terminal. Only the macOS `osx` engine supports it.

```toml
[main.visualizer]
enable = true              # show the spectrum
maxHeight = 10             # maximum rows; 0 fills the available space
fullCharHalfBlock = "▌"    # half-cell character
fullCharFullBlock = "█"    # full-cell character
emptyCharBlock = " "       # empty character
```

The spectrum appears between the lyrics and the song information in the classic layout (the deck layout has the cassette panel instead), with low frequencies at the bottom and high frequencies at the top; foreground and background colors together give it finer amplitude steps.

</details>
<details>
<summary>

### macOS desktop lyrics
</summary>

A native macOS desktop lyrics window with word-by-word highlighting, a spectrum and a customizable look.

```toml
[main.lyric.desktopLyrics]
enable = true                    # show desktop lyrics
showSpectrum = true              # show the spectrum
spectrumStyle = "capsule"        # bar / mirror / capsule / flame / circle / wave / ring / pulse / helix / dna / dualwave

# Appearance
fontName = "PingFang SC"
fontSize = 40.0
textColor = "#FFFFFF"
strokeColor = "#000000"
shadowColor = "#000000"
backgroundColor = "#00000080"    # with transparency
cornerRadius = 16.0
```

- **Word-by-word highlighting** of NetEase YRC lyrics
- **Drag to position**: drag the window to move it; its position is saved
- **Spectrum styles**: 11, from bars to a DNA helix
- **Appearance**: font, colors, outline, shadow, corner radius and transparency

All options are listed in the [in-app changelog](./utils/filex/embed/changelog.md) (in Chinese).

</details>
<details>
<summary>

### Startup animation
</summary>

```toml
[startup]
animation = "deck"
```

- `deck` (default): the cassette turns in from the distance and comes to rest, the wordmark appears letter by letter, and a thin line below fills up as loading progresses
- `sequence`: go-musicfox's sequence animation
- Others: `fade-in`, `rainbow-wave`, `typewriter`, `spinner`, `slide-in`, `glitch`, `matrix-rain`, `particle-burst`

</details>
<details>
<summary>

### Right-click menus
</summary>

Right-clicking an element shows the actions for it:

- **Menu items**: actions for songs, albums and playlists (like, download, collect and more)
- **Status bar**: switch the theme and play mode
- **Breadcrumbs**: menu navigation

Requires mouse support (`enableMouseEvent = true`).

</details>
<details>
<summary>

### Configuration upgrade
</summary>

The `upgrade-config` subcommand adds the settings of a new version to the user's configuration file:

```sh
tapefox upgrade-config
```

- Settings present in the built-in defaults but missing from the user's file are appended to their tables
- Existing values, comments, layout and unknown keys are kept
- The file is replaced atomically and keeps its permissions

Running it after each upgrade is recommended.

</details>
<details>
<summary>

### Key bindings
</summary>

#### In-app keys

##### Built-in actions (fixed)

These actions come from the UI framework; their keys cannot be changed in the configuration.

| Identifier | Action | Default keys |
| :--- | :--- | :--- |
| `rerender` | Redraw the interface | `r`, `R` |
| `moveLeft` | Move left | `h`, `H`, `Left` |
| `moveRight` | Move right | `l`, `L`, `Right` |
| `moveUp` | Move up | `k`, `K`, `Up` |
| `moveDown` | Move down | `j`, `J`, `Down` |
| `moveToTop` | Move to the top | `g` |
| `moveToBottom` | Move to the bottom | `G` |
| `enter` | Enter | `n`, `N`, `Enter` |
| `goBack` | Go back | `b`, `B`, `Esc` |
| `search` | Search the current list | `/`, `／`, `、` |
| `quit` | Quit | `q`, `Q` |

##### Configurable actions

The keys of these actions can be changed in the `[keybindings.app]` table of the configuration.

| Identifier | Action | Default keys |
| :--- | :--- | :--- |
| `help` | Show help | `?`, `？` |
| `pageUp` | Previous page | `Ctrl+u`, `PgUp` |
| `pageDown` | Next page | `Ctrl+d`, `PgDown` |
| `playortoggle` | Play / pause | `Space`, ` `, `　` |
| `toggle` | Toggle playback | *(none)* |
| `previous` | Previous song | `[`, `【` |
| `next` | Next song | `]`, `】` |
| `backwardOneSec` | Seek back 1 second | `x` |
| `backwardFiveSec` | Seek back 5 seconds | `X` |
| `forwardFiveSec` | Seek forward 5 seconds | `v` |
| `forwardTenSec` | Seek forward 10 seconds | `V` |
| `downVolume` | Volume down | `-`, `−`, `ー` |
| `upVolume` | Volume up | `=`, `＝` |
| `switchPlayMode` | Switch play mode | `p` |
| `intelligence` | Heartbeat mode | `P` |
| `clearSongCache` | Clear the music cache | `u`, `U` |
| `logout` | Log out and quit | `w`, `W` |
| `curPlaylist` | Show the current playlist | `c`, `C` |
| `appendSongsToNext` | Play next | `e` |
| `appendSongsAfterCurPlaylist` | Add to the end of the playlist | `E` |
| `delSongFromCurPlaylist` | Remove the selected song from the playlist | `\`, `、` |
| `likePlayingSong` | Like the playing song | `,`, `，` |
| `dislikePlayingSong` | Unlike the playing song | `.`, `。` |
| `trashPlayingSong` | Mark the playing song as disliked | `t` |
| `addPlayingSongToUserPlaylist` | Add the playing song to a playlist | `` ` `` |
| `removePlayingSongFromUserPlaylist` | Remove the playing song from a playlist | `~`, `～` |
| `downloadPlayingSong` | Download the playing song | `d` |
| `downloadPlayingSongLrc` | Download the lyrics of the playing song | `Ctrl+l` |
| `openAlbumOfPlayingSong` | Open the album of the playing song | `a` |
| `openArtistOfPlayingSong` | Open the artist of the playing song | `s` |
| `openPlayingSongInWeb` | Open the playing song on the web | `o` |
| `simiSongsOfPlayingSong` | Songs similar to the playing song | `f` |
| `likeSelectedSong` | Like the selected song | `<`, `〈`, `＜`, `《`, `«` |
| `dislikeSelectedSong` | Unlike the selected song | `>`, `〉`, `＞`, `》`, `»` |
| `trashSelectedSong` | Mark the selected song as disliked | `T` |
| `addSelectedSongToUserPlaylist` | Add the selected song to a playlist | `Tab` |
| `removeSelectedSongFromUserPlaylist` | Remove the selected song from a playlist | `Shift+Tab` |
| `downloadSelectedSong` | Download the selected song | `D` |
| `downloadSelectedSongLrc` | Download the lyrics of the selected song | *(none)* |
| `openAlbumOfSelectedSong` | Open the album of the selected song | `A` |
| `openArtistOfSelectedSong` | Open the artist of the selected song | `S` |
| `openSelectedItemInWeb` | Open the selected song, album or other item on the web | `O` |
| `collectSelectedPlaylist` | Collect the selected playlist | `;`, `:`, `：`, `；` |
| `discollectSelectedPlaylist` | Uncollect the selected playlist | `'`, `"` |
| `subscribeAlbumOfPlayingSong` | Collect the album of the playing song | *(none)* |
| `unsubscribeAlbumOfPlayingSong` | Uncollect the album of the playing song | *(none)* |
| `subscribeArtistOfPlayingSong` | Collect the artist of the playing song | *(none)* |
| `unsubscribeArtistOfPlayingSong` | Uncollect the artist of the playing song | *(none)* |
| `subscribeAlbumOfSelectedSong` | Collect the album of the selected song | *(none)* |
| `unsubscribeAlbumOfSelectedSong` | Uncollect the album of the selected song | *(none)* |
| `subscribeArtistOfSelectedSong` | Collect the artist of the selected song | *(none)* |
| `unsubscribeArtistOfSelectedSong` | Uncollect the artist of the selected song | *(none)* |
| `simiSongsOfSelectedSong` | Songs similar to the selected song | `F` |
| `actionOfSelected` | Action menu for the selection or the playing song | `m` |
| `actionOfPlayingSong` | Action menu for the playing song | `M` |
| `switchTheme` | Switch theme | *(none; available in the right-click menu)* |
| `toggleSortOrder` | Toggle sort order (radio / podcast lists) | `\|` |
| `nowPlaying` | Open / close the now-playing page | `i` |

Notes:

- Named keys are case-insensitive, so `shift+tab` equals `Shift+Tab`; character keys are case-sensitive, so `a` and `A` differ.
- Binding one key to several actions is undefined; the binding read last wins.
- The keys of built-in actions cannot be changed, and the keys they use cannot be bound to other actions.
- Actions on songs, playlists and similar items are also available from the `actionOfSelected` and `actionOfPlayingSong` menus.

Example:

```toml
[keybindings]
# Drop all default bindings (every action then needs a key of its own)
useDefaultKeyBindings = false

[keybindings.app]
# Next song on Alt+N
next = "alt+n"
# Ctrl+H for help, in addition to ?
help = ["?", "ctrl+h"]
# Unbind P from heartbeat mode
intelligence = ""
# Show the current playlist with Ctrl+P
curPlaylist = "ctrl+p"
```

#### Global hotkeys

No global hotkeys are set by default. They are configured in `[keybindings.global]`:

```toml
[keybindings.global]
# key = action
"ctrl+shift+space" = "toggle"
```

Release builds have global hotkeys enabled:

- **Linux**: keys are observed through the X11 RECORD extension, which needs an X11 session; under Wayland they usually do not work.
- **macOS**: the terminal in use needs the Accessibility permission in System Settings → Privacy & Security → Accessibility.
- **Windows**: no setup is needed.

</details>
<details>
<summary>

### Mouse support
</summary>

The following mouse actions are supported:

- **Click**: select menu items, press buttons
- **Double-click**: enter a submenu or start playback
- **Scroll wheel**: scroll lists
- **Hover**: highlight interactive elements
- **Right-click**: open the context menu

```toml
[main]
enableMouseEvent = true  # enable mouse support
```

See [mouse support](./docs/mouse_support.md) (in Chinese) for details.

</details>
<details>
<summary>

### Custom share templates
</summary>

Shared text is produced from templates written in Go's [text/template](https://pkg.go.dev/text/template) syntax, which supports conditionals and more.

#### Fields

| Group | Fields |
| :--- | :--- |
| `Song` | `SongId`: song ID <br> `SongName`: song name <br> `SongArtists`: song artists <br> `SongUrl`: song URL |
| `Album` | `AlbumId`: album ID <br> `AlbumName`: album name <br> `AlbumUrl`: album URL <br> `AlbumArtists`: album artists |
| `Artist` | `ArtistId`: artist ID <br> `ArtistName`: artist name <br> `ArtistUrl`: artist URL |
| `Playlist` | `PlaylistId`: playlist ID <br> `PlaylistName`: playlist name <br> `PlaylistUrl`: playlist URL |
| `User` | `UserID`: user ID <br> `UserName`: user nickname <br> `UserUrl`: user URL |
| `DjRadio` | `DjRadioId`: podcast ID <br> `DjRadioName`: podcast name <br> `DjRadioUrl`: podcast URL |
| `Episode` | `EpisodeId`: episode ID <br> `EpisodeName`: episode name <br> `EpisodeUrl`: episode URL |

A podcast episode (`episode`) has all the fields of `Episode`, `Song` and `Album`; `EpisodeName` and `SongName` hold the same value.

#### Templates

Each kind of shared item has its own fields. The template names are fixed.

| Template | Shares | Field groups |
| :--- | :--- | :--- |
| `song` | a song | `Song`, `Album` |
| `album` | an album | `Album` |
| `artist` | an artist | `Artist` |
| `playlist` | a playlist | `Playlist`, `User` |
| `user` | a user | `User` |
| `djRadio` | a podcast | `DjRadio`, `User` |
| `episode` | a podcast episode | `Episode`, `DjRadio`, `User`, `Song`, `Album` |

#### Example

```toml
[share]
song = "Listening to {{.SongName}}{{if .SongArtists}} by {{.SongArtists}}{{end}}: {{.SongUrl}}"
```

</details>

## Development

Requirements:

- Go 1.26 or later
- On Linux, the FLAC and ALSA development packages:
  - Debian / Ubuntu: `sudo apt install build-essential libflac-dev libasound2-dev`
  - Fedora: `sudo dnf install flac-devel alsa-lib-devel`
  - Arch Linux: `sudo pacman -S flac alsa-lib`

Build and install:

```sh
git clone https://github.com/Egdon/tapefox
cd tapefox
make           # builds bin/tapefox
make install   # installs to $GOPATH/bin
```

Run the tests (`MUSICFOX_ROOT` points at a temporary directory, leaving the local data alone):

```sh
MUSICFOX_ROOT=$(mktemp -d) make test
```

Build the Linux and Windows release packages (requires podman or docker; the output goes to `dist/`):

```sh
make release-snapshot
```

On Fedora, `./hack/rpm.sh` (requires `rpm-build`) builds an RPM for the local system. The project layout, the release process and the conventions are described in [AGENTS.md](./AGENTS.md) (in Chinese).

## Issues and contributions

Report problems with tapefox in [Issues](https://github.com/Egdon/tapefox/issues), with the output of `tapefox --version`, the operating system and terminal, the steps to reproduce and any relevant logs.

Keep pull requests focused and describe their purpose and impact; commit messages follow [Conventional Commits](https://www.conventionalcommits.org/). Do not commit build output, private configuration or tokens.

## Related projects

- [go-musicfox/bubbletea](https://github.com/go-musicfox/bubbletea): a customized [bubbletea](https://github.com/charmbracelet/bubbletea)
- [go-musicfox/netease-music](https://github.com/go-musicfox/netease-music): derived from [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo), without the API server, keeping service and util as a standalone Go package

## Acknowledgements

tapefox builds on:

- [go-musicfox](https://github.com/go-musicfox/go-musicfox) (GPL-3.0): the upstream project, whose updates are merged regularly. Thanks to anhoder and all contributors.
- [foxful-cli](https://github.com/anhoder/foxful-cli): the TUI framework of go-musicfox. tapefox maintains a modified copy in `third_party/foxful-cli`, with its origin and base version recorded in `UPSTREAM.md`; the upstream repository has no license file at present.

And the projects go-musicfox thanks, with their contributors (among others):

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [beep](https://github.com/faiface/beep)
- [musicbox](https://github.com/darknessomi/musicbox)
- [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)
- [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo)
- [gcli](https://github.com/gookit/gcli)

Thanks to the [LINUX DO](https://linux.do) community for its support.

## Disclaimer

- tapefox is an independent fork of go-musicfox, not an official version.
- This project owns no audio, lyrics, cover art or other copyrighted data of any music platform.
- Users are responsible for checking the laws where they live and the terms of the music platforms, and use the software at their own risk.
