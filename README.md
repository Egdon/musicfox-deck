# musicfox-deck

**磁带卡座风格的命令行网易云音乐。** musicfox-deck 是 [go-musicfox](https://github.com/go-musicfox/go-musicfox) 的二次开发版本：主界面左边是曲目表，右边一盘随音乐转动的 3D 磁带；播放页把磁带放上舞台；页面切换、菜单、弹窗和选中项都有克制而顺滑的动效。其余功能与原版一致：各种音质级别、UnblockNeteaseMusic、Last.fm、MPRIS、DLNA 投送等。

<p><img src="docs/assets/deck-panel.gif" alt="主界面：曲目表与磁带面板" width="880"/></p>

> 配置和数据目录与原版共用（`go-musicfox`），原版的登录状态、配置和缓存可以直接使用。

------------------------------
## 特色

- **3D 磁带**：透视光线投射实时渲染。在 Ghostty、kitty、foot 等终端里用八分块字符绘制（每格 2×4 个子像素），其他终端自动改用六分块或四分块。播放时缓慢漂浮，鼠标移上去会跟着倾斜，切歌时立体翻面，暂停时落定；两个带盘的带层随播放进度此消彼长。
- **主界面**：曲目表（序号、歌名、歌手、时长）占满高度，右侧磁带面板显示正在播放、逐字歌词和接下来播放。点磁带进入播放页，点接下来播放里的歌直接播放。窗口窄于 110 列时，面板收成曲目表下方的播放条。
- **播放页**：磁带在左侧舞台上随窗口缩放，右侧是编号信息栏；窄窗口居中排成海报，矮窗口横排。进出播放页时，面板上的小磁带直接放大到舞台，或缩回面板。
- **动效**：页面切换、菜单进出、弹窗、通知、选中项和启动页都有 150–250ms 的过渡，不回弹；可以调成只淡入淡出，或完全关闭。
- **Warm Night 主题**：暖炭灰底配米灰文字，信号橙只标出正在发生的事，为夜间使用设计。

## 预览

| 播放页：磁带漂浮与逐字歌词 | 进入播放页：磁带放大到舞台 |
|---|---|
| <img src="docs/assets/now-playing.gif" alt="播放页" width="440"/> | <img src="docs/assets/morph.gif" alt="进入播放页" width="440"/> |
| **切歌：磁带立体翻面** | **启动：磁带从景深处转入** |
| <img src="docs/assets/flip.gif" alt="切歌翻面" width="440"/> | <img src="docs/assets/startup.gif" alt="启动动画" width="440"/> |
| **菜单进出：逐行滑入** | **窄窗口：底部播放条与海报式播放页** |
| <img src="docs/assets/menus.gif" alt="菜单过渡" width="440"/> | <img src="docs/assets/narrow.gif" alt="窄窗口" width="440"/> |

## 安装

### 下载

前往 [Releases](https://github.com/Egdon/musicfox-deck/releases) 下载对应平台的压缩包，或 deb、rpm、apk、Arch 安装包，运行其中的 `musicfox`。

### 手动编译

需要 Go 1.26 及以上版本。Linux 还需要 FLAC 和 ALSA 的开发包：

* APT（Debian、Ubuntu）：`sudo apt install build-essential libflac-dev libasound2-dev`
* DNF（Fedora）：`sudo dnf install flac-devel alsa-lib-devel`
* pacman（Arch）：`sudo pacman -S flac`

```sh
$ git clone https://github.com/Egdon/musicfox-deck
$ cd musicfox-deck
$ make           # 编译到 bin 目录
$ make install   # 安装到 $GOPATH/bin
```

### Fedora 等 RPM 发行版

在源码目录执行 `./hack/rpm.sh`（需要 `rpm-build`），安装包生成在 `dist/` 下，然后：

```sh
$ sudo dnf install ./dist/musicfox-deck-*.rpm
```

运行依赖只有 `flac-libs`、`alsa-lib` 和 glibc，由安装包自动声明。网页登录窗口需要 WebKitGTK（`webkitgtk6.0` 或 `webkit2gtk4.1`，作为推荐依赖），没有时仍可扫码登录。

> Homebrew、AUR、Scoop、Flatpak 等渠道安装的是原版 go-musicfox，见[原版 README](https://github.com/go-musicfox/go-musicfox#安装)。

## 使用

```sh
$ musicfox
```

<details>
<summary>

### 注意事项
</summary>

- **请务必使用等宽字体，或将配置项 `doubleColumn` 设为 `false`，否则双列显示排版可能会混乱**
- **如果在使用时出现莫名奇妙的光标移动、切歌或暂停等现象，请将配置项 `enableMouseEvent` 设置为 `false`**
- **本应用不对 macOS 原生终端和 Windows 的命令提示符（CMD）做兼容处理（[#99](https://github.com/go-musicfox/go-musicfox/issues/99)）**
  > macOS 用户推荐使用 [iTerm2](https://iterm2.com/) 或 [Kitty](https://sw.kovidgoyal.net/kitty/)
  >
  > Linux 用户推荐使用 [Kitty](https://sw.kovidgoyal.net/kitty/)
  >
  > Windows 用户推荐使用 [Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)，使用体验更佳
- 如果在执行文件时遇到以下错误，说明你的操作系统内不包含 `libFLAC.so.8`：
  ```
  ./musicfox: error while loading shared libraries: libFLAC.so.8: cannot open shared object file: No such file or directory
  ```
  例如 Ubuntu 23.10 及它的衍生版系列，`libFLAC.so.12` 已经将 `libFLAC.so.8` 替换。

  遇到这种问题，你可以：
  * 找到已安装的新版 `libFLAC.so`，将其软链为`libFLAC.so.8`: `ln -s /xxx/libFLAC.so /xxx/libFLAC.so.8` （**推荐**）
  * 自行安装 `libflac8` （不推荐）
  * 参照[手动编译](#手动编译)一节自行编译。

  > 这里之所以使用 FLAC8，主要是为了兼容大部分系统，因为FLAC是向前兼容的（也就是说 `≥ 8` 的FLAC都可以使用）
- wsl 环境下使用 beep 须安装 `libasound2-plugins`，见 [issues](https://github.com/microsoft/wslg/issues/864)
- XDG 支持：完整的 XDG 支持，部分文件路径变更，见 [XDG 支持说明](./docs/xdg_support.md)
- 配置文件格式：使用 [TOML](https://toml.io/) 格式

</details>
<details>
<summary>

### DLNA 投送
</summary>

DLNA 引擎允许你将音乐投送到兼容 DLNA 的设备（如智能电视、音响、游戏机等）。

1. **获取设备 URL**
   - 安装 `gupnp-tools`：`brew install gupnp-tools`（macOS）或 `sudo apt install gupnp-tools`（Linux）
   - 运行 `gupnp-universal-cp`，扫描并复制设备的描述 URL（形如 `http://192.168.1.100:49152/description.xml`）

2. **配置**
   在 `config.toml` 中设置：
   ```toml
   [player]
   engine = "dlna"

   [player.dlna]
   deviceUrl = "http://你的设备IP:端口/description.xml"
   localIP = "本机局域网IP"  # 例如 192.168.1.50
   ```

3. **播放**
   启动应用后，选择歌曲即可自动投送到 DLNA 设备。


</details>
<details>
<summary>
引入完整的主题配置系统，支持运行时切换、自定义配色和多种内置主题。
### 主题系统
</summary>

完整的主题配置系统，支持运行时切换、自定义配色和多种内置主题。

#### 内置主题

- **Default** - 经典浅色主题
- **Transparent** - 透明主题，适配终端背景
- **Dark Transparent** - 深色透明主题
- **Dracula** - 流行的 Dracula 配色
- **Nord** - 北欧冷色调
- **Gruvbox** - 复古暖色调
- **Warm Night**（musicfox-deck 默认）- 暖夜：Braun 式暖炭灰底配米灰文字，信号橙只标出正在发生的事，灰绿表示已开启的开关；为夜间使用设计，与播放页共用一套色板

#### 使用方式

1. **配置文件设置**
   ```toml
   [theme]
   activeTheme = "Dracula"  # 启动时加载的主题
   ```

2. **运行时切换**
   - 通过右键菜单选择"切换主题"
   - 或在配置中绑定快捷键：`switchTheme = "t"`

3. **自定义主题**
   - 在用户配置目录下创建 `themes/` 子目录
   - 每个主题一个独立的 `.toml` 文件（例如：
     - macOS: `~/Library/Application Support/go-musicfox/themes/mytheme.toml`
     - Linux: `~/.config/go-musicfox/themes/mytheme.toml`
     - Windows: `%AppData%\go-musicfox\themes\mytheme.toml`）
   - 主题文件必须包含 `name` 字段和 `[dark]` 或 `[light]` 配置表
   - 详细配置参见 [内置主题示例](./utils/filex/embed/themes/)

</details>
<details>
<summary>

### 磁带卡座界面
</summary>

`[theme] layout = "deck"`（默认）启用，改成 `"classic"` 恢复原版布局。

- **曲目表**：序号、歌名、歌手、时长四列，上方是分节标题，如 `07 ── 飙升榜  [更新27首] ……… 100 首`。在曲目间移动时，选中标记会滑过去。
- **磁带面板**：窗口不小于 110 列 × 24 行时显示在右侧：迷你磁带、歌名、逐字歌词和接下来播放。点磁带进入播放页，点接下来播放里的歌直接播放，滚轮调音量。
- **底部播放条**：更窄或更矮的窗口里，面板收成曲目表下方的一条：小磁带、歌名与歌手、正在唱的歌词，右侧是播放位置、播放模式和音量。

相关配置：

```toml
[main]
motion = "full"        # 动效：full（默认）/ subtle（只淡入淡出）/ off（关闭）
blockGlyphs = "auto"   # 磁带用的块字符：auto / octant / sextant / quadrant / half

[theme]
layout = "deck"        # deck / classic
activeTheme = "Warm Night"

[startup]
animation = "deck"     # 磁带从景深处转入，字标逐字浮现
```

`blockGlyphs = "auto"` 时按终端自动选择：Ghostty、kitty、foot 和较新的 VTE 终端用八分块，WezTerm、iTerm2、Alacritty、Konsole、Windows Terminal 用六分块，tmux 和其他终端用四分块。通过 SSH 使用时，建议把 `motion` 设为 `subtle`。

</details>
<details>
<summary>

### 播放页
</summary>

在主界面按 `i` 或点击磁带进入，`i` 或 `Esc` 返回。

- **宽窗口**（不小于 100 列 × 28 行）：左侧舞台上是随窗口缩放的大磁带；右侧是编号信息栏，`02 TRACK` 列出歌名、歌手、专辑和音质，`03 LYRICS` 是逐字歌词；下方是七段数码管计数器和 `04 UP NEXT`。
- **窄窗口**：磁带居中排成海报，下面是歌名、歌手和正在唱的几行歌词，再下面是时间和进度条。
- **矮窗口**：磁带在左，右侧只放曲目、时间和歌词。
- **磁带**：播放时缓慢漂浮，鼠标移上去会跟着倾斜，切歌时立体翻面，暂停时落定；标签上印着 A/B 面和书写横线。
- **逐字歌词**：有逐字时间轴（YRC）时，已唱的部分逐字亮起；只有行时间时按行高亮。
- **其他**：页头指示灯常驻显示随机、循环、单曲、红心、歌词、高音质等状态；录像机风格的 OSD 显示播放状态和音量；按 `L` 切换全屏歌词；停止且队列为空时显示 NO SIGNAL 蓝屏。
- **页内按键**：`←` / `→` 快退 / 快进 5 秒，`↑` / `↓` 调音量；其余快捷键沿用你的绑定，作用于正在播放的歌曲。
- **鼠标**：点进度条跳转，点底部按键栏控制播放，点接下来播放里的歌直接播放，滚轮调音量。

主界面调音量、暂停、快进、切换模式时，右上角也会短暂显示同样风格的 OSD 提示（`[main.notification] inApp = false` 时关闭）。

</details>
<details>
<summary>

终端内实时显示音频频谱（仅 macOS `osx` 播放引擎支持）。
</summary>

终端 TUI 中的实时音频频谱可视化（仅 macOS `osx` 播放引擎支持）。

#### 配置

```toml
[main.visualizer]
enable = true              # 启用频谱显示
maxHeight = 10             # 最大高度（行数），0 为自动填充
fullCharHalfBlock = "▌"    # 半填充字符
fullCharFullBlock = "█"    # 全填充字符
emptyCharBlock = " "       # 空白字符
```

频谱会显示在歌词和歌曲信息之间，低频在底部，高频在顶部，使用前景/背景双色渐变提供更高精度的幅度表现。

</details>
<details>
<summary>

### macOS 桌面歌词
</summary>
macOS 原生桌面歌词窗口，支持 YRC 逐字高亮、频谱可视化和完整的视觉定制。

#### 配置

```toml
[main.lyric.desktopLyrics]
enable = true                    # 启用桌面歌词
showSpectrum = true              # 显示频谱
spectrumStyle = "capsule"        # 频谱样式：bar/mirror/capsule/flame/circle/wave/ring/pulse/helix/dna/dualwave

# 外观定制
fontName = "PingFang SC"
fontSize = 40.0
textColor = "#FFFFFF"
strokeColor = "#000000"
shadowColor = "#000000"
backgroundColor = "#00000080"    # 支持透明度
cornerRadius = 16.0
```

#### 功能

- **YRC 逐字高亮** - 支持网易云的逐字歌词格式
- **拖拽定位** - 按住窗口可自由移动，位置自动保存
- **11种频谱样式** - 从经典柱状到创意 DNA 螺旋
- **完整视觉定制** - 字体、颜色、阴影、圆角、透明度

详细配置选项参见 [changelog](./utils/filex/embed/changelog.md)。

</details>
<details>
<summary>

### 启动动画
</summary>

- `deck`（默认）：磁带从景深处转入并落定，字标逐字浮现，下方细线随加载填满
- `sequence`：原版的序列动画
- 其他：`fade-in`、`rainbow-wave`、`typewriter`、`spinner`、`slide-in`、`glitch`、`matrix-rain`、`particle-burst`

```toml
[startup]
animation = "deck"
```

</details>
<details>
<summary>

### 右键菜单
</summary>

完整的上下文菜单系统，右键点击不同元素可触发相应操作：

- **菜单项右键** - 显示歌曲/专辑/歌单相关操作（喜欢、下载、收藏等）
- **状态栏右键** - 快速切换主题、播放模式
- **面包屑右键** - 菜单导航操作

可在配置中设置 `enableMouseEvent = true` 以启用完整鼠标支持。

</details>
<details>
<summary>

### 配置升级
</summary>

`upgrade-config` 子命令用于同步新版本的配置项：

```sh
musicfox upgrade-config
```

该命令会：
- 将内嵌默认配置中用户文件缺失的叶子项追加到对应 table
- 保留用户已有的值、注释、布局和未知键
- 以原子操作写回，保持文件权限

建议在升级到新版本后运行一次。

</details>
<details>
<summary>

### 快捷键
</summary>

#### 应用内快捷键

##### 不可自定义操作 (内置)

这些是程序核心框架提供的基础操作，其快捷键通常**无法**通过配置文件修改。

| 配置标识符        | 操作说明        | 默认按键                                      |
| :---------------- | :-------------- | :-------------------------------------------- |
| `rerender`        | 重新渲染UI      | `r`, `R`                                      |
| `moveLeft`        | 左              | `h`, `H`, `Left`                              |
| `moveRight`       | 右              | `l`, `L`, `Right`                             |
| `moveUp`          | 上              | `k`, `K`, `Up`                                |
| `moveDown`        | 下              | `j`, `J`, `Down`                              |
| `moveToTop`       | 上移到顶部      | `g`                                           |
| `moveToBottom`    | 下移到底部      | `G`                                           |
| `enter`           | 进入            | `n`, `N`, `Enter`                             |
| `goBack`          | 返回上一级      | `b`, `B`, `Esc`                               |
| `search`          | 搜索当前列表    | `/`, `／`, `、`                             |
| `quit`            | 退出            | `q`, `Q`                                      |

##### 可自定义操作

你可以在配置文件的 `[keybindings]` 部分修改这些操作的快捷键。

| 配置标识符                          | 操作说明                      | 默认按键                                        |
| :---------------------------------- | :---------------------------- | :---------------------------------------------- |
| `help`                              | 帮助信息                      | `?`, `？`                                     |
| `pageUp`                            | 上一页                        | `Ctrl+u`, `PgUp`                                |
| `pageDown`                          | 下一页                        | `Ctrl+d`, `PgDown`                              |
| `playortoggle`                      | 播放/暂停                     | `Space`, ` `, `　`                           |
| `toggle`                            | 切换播放状态                  | *(无)*                                          |
| `previous`                          | 上一首                        | `[`, `【`                                     |
| `next`                              | 下一首                        | `]`, `】`                                     |
| `backwardOneSec`                    | 快退1秒                       | `x`                                             |
| `backwardFiveSec`                   | 快退5秒                       | `X`                                             |
| `forwardFiveSec`                    | 快进5秒                       | `v`                                             |
| `forwardTenSec`                     | 快进10秒                      | `V`                                             |
| `downVolume`                        | 减小音量                      | `-`, `−`, `ー`                                |
| `upVolume`                          | 加大音量                      | `=`, `＝`                                     |
| `switchPlayMode`                    | 切换播放模式                  | `p`                                             |
| `intelligence`                      | 心动模式                      | `P`                                             |
| `clearSongCache`                    | 清除音乐缓存                  | `u`, `U`                                        |
| `logout`                            | 注销并退出                    | `w`, `W`                                        |
| `curPlaylist`                       | 显示当前播放列表              | `c`, `C`                                        |
| `appendSongsToNext`                 | 添加为下一曲播放              | `e`                                             |
| `appendSongsAfterCurPlaylist`       | 添加到播放列表末尾            | `E`                                             |
| `delSongFromCurPlaylist`            | 从播放列表删除选中歌曲        | `\`, `、`                                     |
| `likePlayingSong`                   | 喜欢播放中歌曲                | `,`, `，`                                     |
| `dislikePlayingSong`                | 取消喜欢播放中歌曲            | `.`, `。`                                     |
| `trashPlayingSong`                  | 标记播放中歌曲为不喜欢        | `t`                                             |
| `addPlayingSongToUserPlaylist`      | 将播放中歌曲加入歌单          | `` ` ``                                         |
| `removePlayingSongFromUserPlaylist` | 将播放歌曲从歌单中删除        | `~`, `～`                                     |
| `downloadPlayingSong`               | 下载播放中歌曲                | `d`                                             |
| `downloadPlayingSongLrc`            | 下载当前播放音乐歌词          | `Ctrl+l`                                        |
| `openAlbumOfPlayingSong`            | 播放中歌曲的所属专辑          | `a`                                             |
| `openArtistOfPlayingSong`           | 播放中歌曲的所属歌手          | `s`                                             |
| `openPlayingSongInWeb`              | 网页打开播放中歌曲            | `o`                                             |
| `simiSongsOfPlayingSong`            | 与播放中歌曲相似的歌曲        | `f`                                             |
| `likeSelectedSong`                  | 喜欢选中歌曲                  | `<`, `〈`, `＜`, `《`, `«`                 |
| `dislikeSelectedSong`               | 取消喜欢选中歌曲              | `>`, `〉`, `＞`, `》`, `»`                 |
| `trashSelectedSong`                 | 标记选中歌曲为不喜欢          | `T`                                             |
| `addSelectedSongToUserPlaylist`     | 将选中歌曲加入歌单            | `Tab`                                           |
| `removeSelectedSongFromUserPlaylist`  | 将选中歌曲从歌单中删除        | `Shift+Tab`                                     |
| `downloadSelectedSong`              | 下载选中歌曲                  | `D`                                             |
| `downloadSelectedSongLrc`           | 下载选中歌曲的歌词            |                                                 |
| `openAlbumOfSelectedSong`           | 选中歌曲的所属专辑            | `A`                                             |
| `openArtistOfSelectedSong`          | 选中歌曲的所属歌手            | `S`                                             |
| `openSelectedItemInWeb`             | 网页打开选中歌曲/专辑...      | `O`                                             |
| `collectSelectedPlaylist`           | 收藏选中歌单                  | `;`, `:`, `：`, `；`                        |
| `discollectSelectedPlaylist`        | 取消收藏选中歌单              | `'`, `"`                                        |
| `subscribeAlbumOfPlayingSong`       | 收藏播放中歌曲的专辑          | *(无)*                                          |
| `unsubscribeAlbumOfPlayingSong`     | 取消收藏播放中歌曲的专辑      | *(无)*                                          |
| `subscribeArtistOfPlayingSong`      | 收藏播放中歌曲的歌手          | *(无)*                                          |
| `unsubscribeArtistOfPlayingSong`    | 取消收藏播放中歌曲的歌手      | *(无)*                                          |
| `subscribeAlbumOfSelectedSong`      | 收藏选中歌曲的专辑            | *(无)*                                          |
| `unsubscribeAlbumOfSelectedSong`    | 取消收藏选中歌曲的专辑        | *(无)*                                          |
| `subscribeArtistOfSelectedSong`     | 收藏选中歌曲的歌手            | *(无)*                                          |
| `unsubscribeArtistOfSelectedSong`   | 取消收藏选中歌曲的歌手        | *(无)*                                          |
| `simiSongsOfSelectedSong`           | 与选中歌曲相似的歌曲          | `F`                                             |
| `actionOfSelected`                  | 对于选中项或当前播放的操作    | `m`                                          |
| `actionOfPlayingSong`               | 对于当前播放的操作            | `M`                                          |
| `switchTheme`                       | 切换主题样式                  | *(无，可通过右键菜单触发)*                      |
| `toggleSortOrder`                   | 切换排序顺序（电台/播客列表） | `|`                                          |
| `nowPlaying`                        | 进入/退出播放页               | `i`                                             |

注意：
- 非字符快捷键大小写不敏感，如 `shift+tab` 等同 `Shift+Tab`，但 `a` 与 `A` 不同
- 多次绑定同一个键的行为是未定义的，以程序最后读取的为准
- [不可自定义操作](#不可自定义操作-内置) 不可自定义且其使用的键也不可用于自定义
- 与歌曲、歌单等存在关联的操作现已默认添加至 `actionOfSelected` 及 `actionOfPlayingSong`


示例配置：
```toml
[keybindings]
# 取消所有默认键绑定（须自定义键以确保正常使用）
useDefaultKeyBindings = false

# 应用内快捷键
[keybindings.app]
# 将“下一首”改为 Alt+N
next = "alt+n"

# 为“帮助”额外添加 Ctrl+H
help = ["?","ctrl+h"]

# 取消“心动模式”的默认 P 键
intelligence = ""

# 将“显示当前播放列表”改为 Ctrl+P
curPlaylist = "ctrl+p"
```

#### 全局快捷键

默认不设置任何全局快捷键，如果需要请在配置文件中的`global_hotkey`下进行配置，例如：

```toml
[keybindings.global]
# 格式：键=功能
"ctrl+shift+space" = "toggle"
```

> 因为Linux下开启全局快捷键需要安装比较多的依赖，可能你并不需要这个功能，所以Releases中的Linux二进制文件是不支持全局快捷键的
>
> 如果需要开启，请安装[依赖](https://github.com/go-vgo/robotgo#requirements)后手动进行编译:
>
> ```shell
> BUILD_TAGS=enable_global_hotkey make build
> ```

</details>
<details>
<summary>


### 鼠标支持
</summary>

完整的鼠标交互支持：

- **点击** - 选择菜单项、点击按钮
- **双击** - 快速进入子菜单或播放
- **滚轮** - 上下滚动列表
- **悬停** - 高亮显示可交互元素
- **右键菜单** - 上下文相关的操作菜单

**配置**：
```toml
[main]
enableMouseEvent = true  # 启用鼠标支持
```

详见 [鼠标支持说明](./docs/mouse_support.md)

</details>
<details>
<summary>

### 自定义分享模板
</summary>

分享文本的内容是完全可定制的，它由一套模板系统驱动（使用 [text/template](https://pkg.go.dev/text/template) ，支持条件判断等高级用法）。

#### 字段分组详解

| 字段分组   | 包含字段                                                                                                   |
| :--------- | :--------------------------------------------------------------------------------------------------------- |
| `Song`     | `SongId`: 歌曲ID <br> `SongName`: 歌曲名称 <br> `SongArtists`: 歌曲的艺术家名 <br> `SongUrl`: 歌曲链接     |
| `Album`    | `AlbumId`: 专辑ID <br> `AlbumName`: 专辑名称 <br> `AlbumUrl`: 专辑链接 <br> `AlbumArtists`: 专辑的艺术家名 |
| `Artist`   | `ArtistId`: 歌手ID <br> `ArtistName`: 歌手姓名 <br> `ArtistUrl`: 歌手链接                                  |
| `Playlist` | `PlaylistId`: 歌单ID <br> `PlaylistName`: 歌单名称 <br> `PlaylistUrl`: 歌单链接                            |
| `User`     | `UserID`: 用户ID <br> `UserName`: 用户昵称 <br> `UserUrl`: 用户链接                                        |
| `DjRadio`  | `DjRadioId`: 播客电台ID <br> `DjRadioName`: 播客电台名称 <br> `DjRadioUrl`: 播客电台链接                   |
| `Episode`  | `EpisodeId`: 播客节目ID <br> `EpisodeName`: 播客节目名称 <br> `EpisodeUrl`: 播客节目链接                   |

**注意：**
*   对于 `episode`（播客节目）类型，由于其本质也是一首“歌曲”，因此它同时拥有 `Episode`、`Song` 和 `Album` 的所有字段。其中 `EpisodeName` 和 `SongName` 的值是相同的。

#### 模板类型与可用字段
不同类型的分享项目，其可用的信息字段也不同。例如，分享一首“歌曲”时，您可以同时获取到“歌曲信息”和它所属的“专辑信息”；而分享一个“歌手”时，则只有“歌手信息”。
不支持自定义模板名

| 模板名称   | 说明                       | 可用字段分组                                  |
| :--------- | :------------------------- | :-------------------------------------------- |
| `song`     | 用于分享一首单曲。         | `Song`, `Album`                               |
| `album`    | 用于分享一张专辑。         | `Album`                                       |
| `artist`   | 用于分享一位歌手。         | `Artist`                                      |
| `playlist` | 用于分享一个歌单。         | `Playlist`, `User`                            |
| `user`     | 用于分享一位用户。         | `User`                                        |
| `djRadio`  | 用于分享一个播客电台。     | `DjRadio`, `User`                             |
| `episode`  | 用于分享播客中的一期节目。 | `Episode`, `DjRadio`, `User`, `Song`, `Album` |

#### 示例

```toml
[share]
song = "分享{{if .SongArtists}}{{.SongArtists}}的{{end}}单曲《{{.SongName}}》: {{.SongUrl}} (来自@网易云音乐)"
```

</details>

## 配置文件

配置文件路径为用户配置目录下的 `config.toml` 文件，详细可参见[配置示例](./utils/filex/embed/config.toml)。

> 用户配置目录路径：
>
> macOS：`$HOME/Library/Application Support/go-musicfox`
>
> Linux：`$XDG_CONFIG_HOME/go-musicfox` 或 `$HOME/.config/go-musicfox`
>
> Windows：`%AppData%\go-musicfox`
>
> 你可以通过设置 `MUSICFOX_ROOT` 环境变量来自定义用户配置的存储位置

## CHANGELOG

See [CHANGELOG.md](./CHANGELOG.md)

## 相关项目

1. [go-musicfox/bubbletea](https://github.com/go-musicfox/bubbletea)：基于 [bubbletea](https://github.com/charmbracelet/bubbletea) 进行部分定制
2. [go-musicfox/netease-music](https://github.com/go-musicfox/netease-music)：fork 自 [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo) ，在原项目的基础上去除 API 功能，只保留 service 和 util 作为一个独立的包，方便在其他 Go 项目中调用

## 致谢

musicfox-deck 建立在以下项目之上：

* [go-musicfox](https://github.com/go-musicfox/go-musicfox)（GPL-3.0）：本项目 fork 自它，并会持续合并上游的更新。感谢 anhoder 和所有贡献者。
* [foxful-cli](https://github.com/anhoder/foxful-cli)：go-musicfox 的 TUI 框架。本项目把它收进 `third_party/foxful-cli` 维护，来源与基线版本见其中的 `UPSTREAM.md`；它的上游仓库目前没有附带许可证文件。

以及原版致谢的项目及其贡献者们（但不限于）：

* [bubbletea](https://github.com/charmbracelet/bubbletea)
* [beep](https://github.com/faiface/beep)
* [musicbox](https://github.com/darknessomi/musicbox)
* [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)
* [NeteaseCloudMusicApiWithGo](https://github.com/sirodeneko/NeteaseCloudMusicApiWithGo)
* [gcli](https://github.com/gookit/gcli)
* ...
