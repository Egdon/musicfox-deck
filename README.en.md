# musicfox-deck

[简体中文](README.md) | **English**

A terminal client for NetEase Cloud Music with a cassette-deck interface, forked from [go-musicfox](https://github.com/go-musicfox/go-musicfox). The main screen pairs a tracklist with a 3D cassette panel rendered in real time, the now-playing page centers on the cassette, and page changes, menus, popups and the selection all have transitions. Everything else matches go-musicfox, including multiple audio qualities, UnblockNeteaseMusic, Last.fm, MPRIS and DLNA casting.

<p><img src="docs/assets/deck-panel.gif" alt="Main screen: tracklist and cassette panel" width="880"/></p>

> The interface is in Simplified Chinese. musicfox-deck shares its configuration and data directory (`go-musicfox`) with go-musicfox, so an existing login, configuration and cache carry over.

## Features

- **3D cassette**: rendered in real time by perspective ray casting. Terminals with octant characters (Ghostty, kitty, foot and others) draw it at 2×4 subpixels per cell; other terminals fall back to sextants or quadrants. While playing, the cassette floats slowly and tilts toward the mouse; it flips over on a track change and comes to rest when paused. The tape on the two reels follows the playback position.
- **Main screen**: the tracklist (number, title, artist, duration) fills the height, and the cassette panel on the right shows the current track, word-by-word lyrics and the play queue. Clicking the cassette opens the now-playing page; clicking a track in the queue plays it. In windows narrower than 110 columns, the panel becomes a player strip below the tracklist.
- **Now-playing page**: wide windows use a stage layout, with the cassette scaled to the window and a numbered information column on the right; narrow windows use a centered poster layout, short windows a landscape layout. Entering and leaving the page, the panel's cassette morphs into the one on the stage and back.
- **Motion**: page changes, menus, popups, notifications, the selection and the startup screen have 150–250 ms transitions without overshoot. Motion can be reduced to fades or turned off.
- **Warm Night theme**: warm grey background and beige text, with signal orange reserved for what is in progress. Designed for use at night.

## Preview

| Now playing: floating cassette and word-by-word lyrics | Entering the page: the cassette grows onto the stage |
|---|---|
| <img src="docs/assets/now-playing.gif" alt="Now-playing page" width="440"/> | <img src="docs/assets/morph.gif" alt="Entering the now-playing page" width="440"/> |
| **Track change: the cassette flips over** | **Startup: the cassette turns in from the distance** |
| <img src="docs/assets/flip.gif" alt="Track change" width="440"/> | <img src="docs/assets/startup.gif" alt="Startup animation" width="440"/> |
| **Menus: rows slide in** | **Narrow window: player strip and poster layout** |
| <img src="docs/assets/menus.gif" alt="Menu transitions" width="440"/> | <img src="docs/assets/narrow.gif" alt="Narrow window" width="440"/> |

## Installation

| Platform | Method |
|---|---|
| macOS (Apple silicon / Intel) | `brew install egdon/tap/musicfox-deck` |
| Windows 10 / 11 (x64 / ARM64) | Download and extract `windows_amd64.zip` or `windows_arm64.zip` |
| Debian / Ubuntu | Download the `.deb`, then `sudo apt install ./musicfox-deck_*_amd64.deb` |
| Fedora / openSUSE | Download the `.rpm`, then `sudo dnf install ./musicfox-deck-*.x86_64.rpm` |
| Arch Linux | Download the `.pkg.tar.zst`, then `sudo pacman -U musicfox-deck-*-x86_64.pkg.tar.zst` |
| Other Linux distributions | Download `linux_amd64.tar.gz` and put `musicfox` on your `PATH` |

Packages and archives are published on the [Releases](https://github.com/Egdon/musicfox-deck/releases/latest) page, each platform for x86_64 (`amd64` or `x86_64` in the file name) and ARM64 (`arm64` or `aarch64`). Once installed, run `musicfox` in a terminal.

### macOS

Homebrew is the recommended method:

```sh
brew install egdon/tap/musicfox-deck
```

To install manually, download `darwin_arm64.tar.gz` (Apple silicon) or `darwin_amd64.tar.gz` (Intel). The binary is not notarized by Apple: a file downloaded with a browser needs its quarantine attribute removed before the first run, or permission in System Settings → Privacy & Security:

```sh
xattr -d com.apple.quarantine ./musicfox
```

Binaries installed by Homebrew carry no quarantine attribute and need no such step. `darwin_*_macapp.zip` holds the .app version, whose notifications show the app icon.

### Windows

Extract the archive and run `musicfox.exe` in [Windows Terminal](https://aka.ms/terminal), or add its directory to your `PATH`. The binary is not signed; started from File Explorer, it may bring up a SmartScreen prompt, where "More info" → "Run anyway" lets it run. If security software reports it, check the file against `SHA256SUMS` on the Releases page.

### Linux

The packages and the archives contain the same binary, which needs:

- glibc 2.34 or later (Ubuntu 22.04, Debian 12, Fedora 35, RHEL 9 and later releases)
- ALSA (`libasound.so.2`)

The deb, rpm and Arch packages declare these dependencies.

Web login needs WebKitGTK 4.1 (Debian / Ubuntu: `libwebkit2gtk-4.1-0`; Fedora: `webkit2gtk4.1`; Arch Linux: `webkit2gtk-4.1`). The deb and rpm packages recommend it; without it, the login page offers QR code login only.

The beep engine needs `libasound2-plugins` under WSL (see [microsoft/wslg#864](https://github.com/microsoft/wslg/issues/864)).

### Terminals and display

The cassette is drawn with Unicode block characters, so its detail depends on the characters a terminal supports. With `[main] blockGlyphs = "auto"` (the default) the characters follow the terminal:

| Terminal | Characters | Subpixels per cell |
|---|---|---|
| Ghostty, kitty, foot, and VTE terminals such as recent GNOME Terminal | Octants | 2×4 |
| WezTerm, iTerm2, Alacritty, Konsole, Windows Terminal | Sextants | 2×3 |
| Inside tmux, and other terminals | Quadrants | 2×2 |

A monospaced font is required. The macOS Terminal app and the Windows Command Prompt (CMD) are not supported. Over SSH, `[main] motion = "subtle"` is recommended.

### Migrating from go-musicfox

musicfox-deck shares the configuration and data directory with go-musicfox, so the login, configuration and cache stay in use. Both install an executable named `musicfox` and cannot be installed side by side: uninstall go-musicfox first (for example `brew uninstall go-musicfox`). Settings that are new in musicfox-deck can be added to an existing configuration with `musicfox upgrade-config`; see [Configuration upgrade](#configuration-upgrade).

### Building from source

Requires Go 1.26 or later. On Linux, the FLAC and ALSA development packages are needed as well:

- Debian / Ubuntu: `sudo apt install build-essential libflac-dev libasound2-dev`
- Fedora: `sudo dnf install flac-devel alsa-lib-devel`
- Arch Linux: `sudo pacman -S flac alsa-lib`

```sh
git clone https://github.com/Egdon/musicfox-deck
cd musicfox-deck
make           # builds bin/musicfox
make install   # installs to $GOPATH/bin
```

On Fedora, `./hack/rpm.sh` (requires `rpm-build`) builds an RPM for the local system into `dist/`.

## Usage

```sh
musicfox
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
musicfox upgrade-config
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

## Configuration

The configuration file is `config.toml` in the user configuration directory; the [default configuration](./utils/filex/embed/config.toml) lists every setting, with comments in Chinese.

| System | User configuration directory |
|---|---|
| macOS | `$HOME/Library/Application Support/go-musicfox` |
| Linux | `$XDG_CONFIG_HOME/go-musicfox` or `$HOME/.config/go-musicfox` |
| Windows | `%AppData%\go-musicfox` |

The `MUSICFOX_ROOT` environment variable selects another directory.

## Changelog

See [CHANGELOG.md](./CHANGELOG.md).

## Related projects

- [go-musicfox/bubbletea](https://github.com/go-musicfox/bubbletea): a customized [bubbletea](https://github.com/charmbracelet/bubbletea)
- [go-musicfox/netease-music](https://github.com/go-musicfox/netease-music): derived from [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo), without the API server, keeping service and util as a standalone Go package

## Acknowledgements

musicfox-deck builds on:

- [go-musicfox](https://github.com/go-musicfox/go-musicfox) (GPL-3.0): the upstream project, whose updates are merged regularly. Thanks to anhoder and all contributors.
- [foxful-cli](https://github.com/anhoder/foxful-cli): the TUI framework of go-musicfox. musicfox-deck maintains a modified copy in `third_party/foxful-cli`, with its origin and base version recorded in `UPSTREAM.md`; the upstream repository has no license file at present.

And the projects go-musicfox thanks, with their contributors (among others):

- [bubbletea](https://github.com/charmbracelet/bubbletea)
- [beep](https://github.com/faiface/beep)
- [musicbox](https://github.com/darknessomi/musicbox)
- [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)
- [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo)
- [gcli](https://github.com/gookit/gcli)
