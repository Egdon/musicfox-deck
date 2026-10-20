# musicfox-deck 项目架构文档

musicfox-deck 是 go-musicfox 的二次开发版本（仓库 `Egdon/musicfox-deck`，见 `types.AppRepo`）。Go 模块路径、导入路径、二进制名 `musicfox` 和用户数据目录 `go-musicfox` 都与上游保持一致，便于合并上游提交、沿用老用户的数据。

## 沟通语言准则

### 重要准则：与用户交流必须使用中文

**所有贡献者、开发者以及 AI 助手在与用户交流时，必须使用中文。**

#### 适用场景
- 代码审查意见和反馈
- Issue 回复和讨论
- Pull Request 描述和评论
- 文档编写（README、CHANGELOG 等除外，代码注释和提交信息仍使用英文）
- AI 助手与用户的对话

#### 例外情况
- **代码注释**：使用英文（保持代码可读性和国际化标准）
- **Git Commit Message**：使用英文（遵循 Conventional Commits 规范）
- **代码变量/函数命名**：使用英文（遵循 Go 语言惯例）
- **API 文档和错误消息**：根据实际情况使用中文或英文

#### 示例

```markdown
✓ 正确示例（中文交流）：
"这个 PR 修复了歌词封面图圆角功能的性能问题，优化后 CPU 占用从 10% 降至 2%。"

✗ 错误示例（英文交流）：
"This PR fixes the performance issue with lyric cover cornerRadius feature..."
```

## 编码行为准则

本准则旨在减少 AI 编码常见的错误。**权衡原则**：这些准则偏向谨慎而非速度，对于简单任务可自行判断。

> 本准则改编自 [andrej-karpathy-skills/CLAUDE.md](https://github.com/forrestchang/andrej-karpathy-skills/blob/main/CLAUDE.md)

### 1. 三思而后行

**不要假设。不要隐藏困惑。明确权衡。**

在实现之前：
- 明确陈述你的假设。如果不确定，先问。
- 如果存在多种解释方案，全部提出——不要默默选择一个。
- 如果存在更简单的方案，指出它。有理由时可以反驳。
- 如果有不清楚的地方，停下来。说出哪里不清楚，然后问。

### 2. 简单性优先

**最小代码解决问题。不要 speculative。**

- 不添加需求之外的功能。
- 一次性使用的代码不抽象。
- 不添加未被请求的"灵活性"或"可配置性"。
- 不处理不可能发生的错误场景。
- 如果写了 200 行而可以用 50 行完成，重写。

问自己："高级工程师会觉得这过于复杂吗？"如果是，简化。

### 3. 精准修改

**只触碰必须改的。只清理自己造成的混乱。**

编辑现有代码时：
- 不"改进"相邻代码、注释或格式。
- 不重构没坏的东西。
- 匹配现有风格，即使你可能用不同方式实现。
- 如果发现无关的死代码，指出它——不要删除它。

当你的修改产生孤儿（不再使用的代码）时：
- 移除因你的修改而不再使用的 import/变量/函数。
- 不要移除预先存在的死代码，除非被要求。

**检验标准**：每一行修改都应该能直接追溯到用户的请求。

### 4. 目标驱动执行

**定义成功标准。循环验证直到完成。**

将任务转化为可验证的目标：

| 模糊表述 | 可验证目标 |
|---------|-----------|
| "添加验证" | "为无效输入编写测试，然后让测试通过" |
| "修复 bug" | "编写能复现问题的测试，然后让测试通过" |
| "重构 X" | "确保重构前后测试都通过" |

对于多步骤任务，简要陈述计划：

```
1. [步骤] → 验证: [检查点]
2. [步骤] → 验证: [检查点]
3. [步骤] → 验证: [检查点]
```

明确的目标让你能独立循环验证。模糊的目标（"让它能工作"）需要不断确认。

---

**这些准则生效的标志**：diff 中不必要的修改更少，因过度复杂而返工更少，澄清问题在错误之前而非错误之后提出。

## 项目概述

go-musicfox 是基于 Go 和 bubbletea 的网易云音乐 TUI 客户端，支持 macOS/Linux/Windows。musicfox-deck 在其上加了磁带卡座界面（3D 磁带渲染器、曲目表与磁带面板、舞台式播放页）和全局动效，默认主题 Warm Night、布局 deck、启动动画 deck。

**技术栈**：
- **UI 框架**：bubbletea + foxful-cli（部分定制）
- **音频处理**：beep、go-mp3、go-flac
- **存储**：BoltDB
- **配置**：TOML + mapstructure；`configs.SetTOMLValue(path, keyPath, value)` 使用 `github.com/pelletier/go-toml/v2/unstable/edit` 按键路径保真编辑既有 TOML。`configs.UpgradeConfig(path)` 将内嵌默认 TOML 中用户文件缺失的叶子项追加到对应 table；`upgrade-config` 子命令调用它。两者均保留无关注释、布局、已有值、未知键与原文件权限，并以原子替换写回。
- **API**：netease-music SDK

**项目结构**：`cmd/` 入口 | `internal/` 核心业务（22个包） | `utils/` 工具 | `configs/` 嵌入式配置 | `third_party/foxful-cli/` 内置 UI 框架

## 核心架构

### 应用入口与初始化

**入口**：`cmd/musicfox.go` → `runtime.Run()` → 加载配置 → 数据迁移 → 启动 TUI

### UI 协调器

**文件**：`internal/ui/netease.go`

核心结构包含：login、search、player、lyricService、trackManager 等组件。

### 核心接口

**Menu 接口**（`internal/ui/menu.go`）：
```go
type Menu interface {
    model.Menu
    IsPlayable() bool
    IsLocatable() bool
}
type SongsMenu interface { Songs() []structs.Song }
type PlaylistsMenu interface { Playlists() []structs.Playlist }
```

**Player 接口**（`internal/player/player.go`）：
```go
type Player interface {
    Play(music URLMusic)
    Pause()/Resume()/Stop()/Toggle()
    Seek(duration time.Duration)
    PassedTime()/PlayedTime() time.Duration
    Volume()/SetVolume()/UpVolume()/DownVolume()
    State() types.State
    Close()
}
```

### 播放引擎

| 引擎 | 平台 | 特点 |
|------|------|------|
| Beep（默认） | 跨平台 | MP3/FLAC/OGG/WAV |
| DLNA | 跨平台 | 设备投送 |
| MPV | 跨平台 | IPC 控制（命令连接带排空 goroutine 防 mpv 发送缓冲堆积，切歌带 file-loaded 超时 watchdog 兜底） |
| MPD | Linux | 远程服务器 |
| AVFoundation | macOS | 原生集成 |
| MediaPlayer | Windows | WinRT API |

### 事件处理

**文件**：`internal/ui/event_handler.go`

支持 40+ 键盘操作、鼠标事件（单击/双击/滚轮/右键）、可配置快捷键。

### 其他模块

- **歌词**：LRC/YRC 格式，支持 smooth/wave/glow 渲染模式；未匹配的云盘歌曲（含旧播放快照）优先通过云盘歌词接口获取内嵌歌词，失败时回退普通歌曲歌词接口
- **播放列表**：列表循环/顺序/单曲循环/随机/无限随机/智能心动模式
- **远程控制**：MPRIS(linux)、Now Playing(macOS)、System Media(Windows)。macOS 侧 `internal/macdriver/mediaplayer` 与 `internal/macdriver/cocoa/unnotifications.go` 的框架加载（Dlopen）失败不 panic 静默降级：MediaPlayer.framework 需 macOS 10.12.2+、UserNotifications.framework 需 10.14+，旧系统上 class 查找得 nil、objc 消息发送安全返回 0，对应功能（Now Playing 远程控制/系统通知）自动禁用而不影响主应用启动
- **状态栏组件**：`model.DefaultStatusBar.Components` 可注入任意 `StatusBarComponent` 到居中区域；展示项只实现 `View`，可点击项额外实现 `InteractiveStatusBarComponent`，以自身局部坐标处理命中和事件。状态栏负责布局、边界与指针，不持有业务回调。播放队列适配器的 `musicfox` 前缀自行打开 `types.AppGithubUrl`（本 fork 的仓库），队列位置与音质文本不触发。
- **存储**：BoltDB，存储用户信息、播放状态、播放列表快照和桌面歌词窗口位置/显示器
- **扫码登录**：二维码登录（`internal/ui/login_qr_page.go`）通过 `internal/ui/qr_login_client.go` 的 `qrGetKey` / `qrCheckStatus` 直接请求网易云接口，而非 `service.LoginQRService`。客户端基于 `github.com/imroc/req/v3` 的 `SetTLSFingerprintChrome()` 模拟 Chrome 的 TLS ClientHello 指纹并附带完整浏览器 Headers，规避服务端对非浏览器客户端的反爬检测（462）。`qrKeyClient` 请求 unikey 时不带 Cookie；轮询用 `newQRLoginClient(cookieJar)` 每次绑定最新的全局 CookieJar（`neteaseutil.GetGlobalCookieJar()`），避免登录 Cookie（`MUSIC_U`）写入过期 jar。轮询前 `util.ApplyRequestStrategy` 注入反风控 Cookie。462 拦截为网易云动态风控行为，与 IP/地域/频率相关，并非必现。
- **实时频谱**：仅 macOS `osx` 播放引擎支持；`MTAudioProcessingTap` 经 PureGo 获取 PCM，由 `internal/player/spectrum.go` 异步分析。PCM 更新频谱目标值，`github.com/charmbracelet/harmonica` 的临界阻尼弹簧以每帧推进，避免 PCM 回调间隙使动画停顿。频谱分析与 UI 刷新均使用 `[main].frameRate`。`[main.visualizer]` 支持 `enable`、`maxHeight`（默认 `0` 为不限制；正数限制频谱行数）及 `fullCharHalfBlock`、`fullCharFullBlock`、`emptyCharBlock` 字符配置（各取首个 Unicode 字符，默认分别为 `▌`、`█`、空格）。`SpectrumRenderer` 将相邻频段分组为由低至高的横向进度条（低频在底部），以三态字符显示：满单元、使用前景/背景双色渐变的半单元、无样式空白单元，并以半单元为粒度提供双倍横向幅度分辨率。未限制时频谱会占满歌词与歌曲信息之间的可用行，顶部及歌曲信息前各保留一行空白。
- **主题**：`[dark.app]` / `[light.app]` 的显式非透明 `background` 作为应用、菜单行、Startup 页面（普通内容和全屏特效）及 Search、Login、QR Login、Last.fm 授权/二维码授权/API account 等自定义页面完整窗口的终端单元格背景；未设置或 `transparent` 时 Startup 与自定义页面均不写入背景，Default 与 Transparent 保持终端透明。`[*.highlights]` 不提供菜单背景覆写，只有 `selectedItemBg` 可以覆盖由 Primary 计算出的选中行色带。内置主题状态栏的面包屑与时间字体继承 Default 主题（Warm Night 例外，整条状态栏自行配色）。仅 Transparent 主题通过 `statusBar.nuggetLabelFg` / `nuggetLabelBg` 将 `»` 配置为 Primary 前景、透明背景，并以 `statusBar.breadcrumbHover` 将面包屑 hover 前景设为 Primary；右键菜单 hover 文本也使用 `Primary`。菜单选中项的 `menuSelectedSepLeft` / `menuSelectedSepRight` 配置圆角分隔符；右侧符号优先替换原有末尾填充空格，无可替换空格时增加独立符号宽度，不压缩或截断正常内容。居中菜单始终按标称宽度计算起点，额外圆角单元格不会改变标题列；双列模式下左列额外单元格从右列结构性前导空格中吸收，右列起点保持不变。用户通过快捷键或右键菜单切换主题时，会将名称保真写入 `theme.activeTheme`，下次启动加载该主题；系统外观变化仅切换当前主题的明暗变体。
- **磁带渲染器**：`internal/ui/deck` 包，主界面面板、底部播放条、播放页和启动页共用。
  - **几何与着色**：`geom.go` 解析求交圆角外壳（76×48×9 面单位），`face.go` 程序化生成 A/B 两面（标签印 A/B 面字母与书写横线，细节随尺寸做 LOD），`shade.go` 着色：主光、补光、窗口凹槽视差与内阴影、滑过窗口的条形高光；静止时正面颜色就是色板原色。
  - **渲染**：`render.go` 的 `Renderer.Render(View, Palette, Pose, State)` 透视光线投射，只在轮廓和材质边界超采样，每格做 2-means 双色拟合，按行并行，位姿和状态不变时复用上一帧。`View.Fit` 指定磁带适配的盒子（为空时用整个区域）；`FitAspect` 是磁带恰好填满时盒子的宽高比。只有投影落到的格带 `Cell.Shadow`，调用方据此压暗已有内容而不是覆盖（形变过渡时文字不会被投影抹掉）。
  - **字形分级**：`glyphs.go` 的 Octant（2×4，U+1CD00–1CDE5）、Sextant（2×3，U+1FB00–1FB3B）、Quadrant、Half。`detect.go` 的 `DetectTier` 先看 XTVERSION 回复（tmux/screen 一律四分块），再看环境变量；`[main] blockGlyphs` 非 `auto` 时直接采用（`internal/ui/deck_support.go` 的 `deckTier`）。格子宽高比来自 CSI 16t 回报的像素尺寸，只接受 1.4–3.2（`deckAspect`，tmux 会回报虚假尺寸）。
  - **运动**：`motion.go` 的 `Motion` 负责漂浮（三组周期互不整除的正弦，`FloatAmount` 控制幅度：播放页 1、面板 0.5，暂停时落定）、跟随鼠标倾斜（`Point`/`Leave`，2.5s 无移动回正）、切歌翻面（`Flip`，`FlipDuration` 650ms，转过 90° 时换面）；`Reels` 负责带盘转速与卷带。`PackRadii(progress, sideB)` 给出面积守恒的带层半径。
- **主界面（deck 布局）**：`[theme] layout = "deck"`（默认，`DeckLayout()` 对 `classic` 以外的值都返回 true）时，`internal/commands/netease.go` 设置 `options.ListStyle = model.ListTracklist`、`options.SidePanel = ui.NewDeckPanel(...)`、`options.MenuHeader = ui.MenuHeader`（为分节标题补「首 / 个歌单 / 张专辑 / 位歌手」单位）以及 `options.Motion.GlideSteps = ui.GlideSteps(netease)`。
  - **磁带面板**：`internal/ui/deck_panel.go` 的 `DeckPanel` 实现 `model.SidePanel` 与 `model.SidePanelMouse`。窗口不小于 110 列 × 24 行时是右侧一栏（宽度 0.34w，限制在 36–56），`renderDeckPanel` 依次画 NOW PLAYING 页头、迷你磁带、歌名、逐字歌词和接下来播放；footer 保留 3 行（歌曲信息与进度条）。点磁带进入播放页，点接下来播放第 k 项经 `jump` 前进 k 步（保留洗牌顺序），滚轮调音量。
  - **底部播放条**：更窄或更矮时 `Layout` 返回 `Height`（6 行，28 行以上的窗口 7 行），框架把面板停靠在菜单下方；footer 只保留进度条（`deckDockFooter = 1`），播放条自己显示曲目，因此不会和歌曲信息行重复。`renderDeckDock` 画一条细线，下面是小磁带（宽度不足 48 列时省略）、歌名与歌手、正在唱的歌词，右侧对齐播放位置、播放模式和音量（放不下时先省略右侧）。小磁带只有几行高，按 `deckDockTurn` 转向正面、按 `deckDockGrow` 放大适配盒，否则透视会把它糊成一团。
  - **旧组件**：面板可见时（`main.PanelVisible()`）歌词、封面、频谱渲染器不输出。
- **播放页**：`internal/ui/now_playing_page.go`（状态、输入、过渡）+ `now_playing_draw.go`（纯绘制）+ `now_playing_palette.go`（色板）+ `lyric_view.go`（逐字歌词，与面板共用）。主界面按 `nowPlaying`（默认 `i`）或点磁带进入，`i`/`Esc` 返回。
  - **渲染**：每帧先采集 `npSnapshot`，再由 `npRender` 以纯函数方式画进 `uv.Buffer`。页面自己给每一格上色，不调用 `finishCustomPageView`，是「自定义页面使用 app background」规则的例外。写字经 `npNextCluster` 归一化：Grapheme 与 Wc 宽度不一致的字符簇退回首个 rune 或 `?`，保证两种宽度算法下每行都等于窗口宽度。
  - **布局**：`npLayoutOf` 按窗口选择。`npLayoutWide`（≥100 列且 ≥28 行）：左侧舞台占 56% 宽，右侧编号信息栏 `02 TRACK` / `03 LYRICS`，下方七段计数器与 `04 UP NEXT`；`npLayoutLandscape`（≥100 列但矮）：舞台加 TRACK、时间、LYRICS；`npLayoutPoster`（窄于 100 列）：磁带居中，下面居中排歌名、歌手、几行歌词，再下面是时间与进度条（`npPosterPlan` 规划纵向排布）；`npLayoutLyrics`（`L` 全屏歌词）；`npLayoutCompact`（小于 60×20，只有文字）。`npStage` 给出各布局的磁带区域，形变过渡与鼠标倾斜都以它为准。
  - **色板**：`npPaletteFromTheme` 把当前 StyleSet 的 `Theme()` 映射为 `npPalette` 角色（Bg←`[app].background`/`background`，Surface，Text←`foreground`，Hero←`primary`，Secondary，Rule←`muted`，Accent，Accent2←`success`），缺失或为 `NoColor` 的角色用 `npWarmNight` 补齐，Ghost（未点亮）由 Bg 与 Rule 派生；按 `style.StyleGeneration()` 缓存。`npWarmNight` 必须与 `warm_night.toml` 一致（`TestWarmNightPaletteMatchesTheme`）。
  - **形变过渡**：页面实现 `model.PageTransitionHandler`。从 Main 进入且动效为 full 时，`DeckPanel.CassettePlacement()` 给出面板磁带的格子、适配盒与转向，磁带在 320ms 内（OutCubic）从那里放大到舞台，主界面帧前半程淡出、页面后半程淡入，转向随之消退；退出时用 `App.Snapshot(main)` 取目标帧反向播放，结束时 `adopt` 把磁带运动交还面板并 `SuppressPageTransition`。其他情况由 App 交叉淡化。
  - **帧**：页面没有自己的 tick 循环，经 `a.Animate`（16ms）或 `a.AnimateSlow`（50ms，仅漂浮）请求帧，暂停落定后停止。
  - **按键语义**：复用 `EventHandler.keyToOperateMap`，但都作用于正在播放的歌曲。上一首/下一首先翻面，经 `player.async` 切歌，检测到歌曲变化后翻完，1.5s（`npFlipTimeout`）内没有变化则翻回；打开菜单的操作执行后回 Main；`*Selected*` 与 `enter` 等依赖 Main 选中项的操作忽略。
  - **接下来播放**：`PlaylistManager.Upcoming(n)` 只读预览自动播放顺序（无限随机返回 nil）。列表随机照搬 `NextSong` 读取 `randomOrder[currentPos+1:]`：洗牌顺序走到末尾即停，排在当前歌曲之前的歌不会自动播放，当前歌曲排在末位时预览为空。
- **启动动画**：`[startup] animation = "deck"`（默认）时设置 `options.StartupOptions.View = ui.NewDeckStartup().View`（`internal/ui/deck_startup.go`）：磁带从景深处转入（B 面先朝前）并落定，字标逐字浮现，细线随加载填满；时间轴按 2 秒设计，加载更短时等比加快，`motion = off` 时直接画最后一帧。
- **动效**：`[main] motion = full | subtle | off` 设置 `model.Options.Motion.Level`；`off` 同时打开 `[startup].reducedMotion`。框架负责页面交叉淡化、菜单进出、弹窗与通知、选中项滑动（见「foxful-cli（内置源码）」）；页面内的动画都经 `App.Animate` 请求帧，不要用 `RerenderCmd(true)` 驱动动画。
- **菜单加载**：`baseMenu.LoadsInBackground()` 返回 true，菜单的进入钩子默认在后台拉取数据（见框架改动 17）；钩子要碰界面时覆写为 false（`MenuToPage` 会清封面、`SearchResultMenu` 读搜索页）。`Netease.ToLoginPage` 可在后台调用：只记下回调，焦点与清封面在登录页第一次 `Update` 时完成。
- **界面线程不得等网络**：每帧都会读的状态不能和网络 I/O 共用一把锁。`lyric.Service.SetSong` 在锁外拉歌词（`loadSeq` 丢弃过期结果）；beep 播放器加载新歌时持锁连流，故 `PassedTime`、`PlayedTime`、`Volume`、`CurMusic` 读原子缓存或 `TryLock`，不等锁。
- **切歌**：`Player.PlaySong` 会同步请求网络。上一首/下一首、开始播放和跳到接下来播放都经 `player.async` 在 UI goroutine 之外执行，完成后 `Rerender(false)`；`playSeq` 保证并发切歌时只有最后一次生效。
- **OSD**：`internal/ui/osd.go` 提供录像机风格文案，播放页与主界面共用。主界面 OSD 在 `EventHandler.KeyMsgHandle` 与滚轮调音量后以单个 toast 显示，`[main.notification] inApp = false` 时关闭。`App.Notify` 返回的 ID 只是预测值：新增任何非 OSD 的 `App.Notify` 调用点前必须先调用 `markForeignNotify()`，OSD 只在与其他通知相隔 500ms 以上时原地更新。
- **Warm Night 主题**：`utils/filex/embed/themes/warm_night.toml`，只做深色，`[light]` 镜像 `[dark]`。角色：bg `#151412`、surface `#1D1C19`、text `#CBC4B6`、hero `#DDD6C9`、secondary `#9E978A`、rule `#544F47`、accent `#D2714F`（信号橙）、accent2 `#93AC82`（灰绿）。为夜间使用而设计：不用纯白，彩色格子控制在约 5% 以内，不要把 accent 用于普通文字。markdown 样式为 `builtin:warm_night`。依赖内置 foxful-cli 的改动（见「foxful-cli（内置源码）」）。
- **主题扩展键**（`internal/configs/theme_file.go` 的 `toTheme`，未设置时外观不变）：顶层 `menuSelectedArrow`（选中行标记，必须正好 3 格宽，否则用 `=> `）；`[highlights]` `selectedItemBg`、`menuArrow`、`menuIndex`、`tabBorder`；`[popup]` `actionFocusedFg`、`actionFocusedBg`、`contextMenuItem`；`[notification]` `title`、`message`、`infoIcon`、`successIcon`、`warningIcon`、`errorIcon`；`[statusBar]` `barBg`、`breadcrumbBg`、`timeBg`、`timeIcon`（`""` 去掉时钟图标）；`[app]` `contextMenuBorder`（此前声明了也不生效，现在 Dracula、Nord、Gruvbox、GitHub 的右键菜单边框按其声明着色）。`[*.markdown]` 的值以 `builtin:` 开头时指向 `utils/filex/embed/markdown/<name>.json`：glamour 只读文件，`markdownStyle` 先把它写到缓存目录的 `markdown/` 下。
- **主题色辅助函数**：`internal/ui/theme_colors.go` 的 `themePrimary`、`themeSecondary`、`themeAccent` 读取当前 StyleSet 的 `Theme()`，主题未设置时回退旧的 `[theme].primaryColor`；新界面代码应使用它们而不是 `util.GetPrimaryColor()`。表单页提示色用 `tipError`、`tipSuccess`、`tipInfo`，读取主题 `[app]` 的 `loginError`、`loginSuccess`、`authStatus`，默认 ANSI 亮红、亮绿、亮蓝。Login、Search 页的输入框和按钮在构造时主题尚未生效，进入页面时按当前主题刷新样式。
- **播放栏**：`SongInfoRenderer.View` 生成片段（模式、`VOL n`、`▶`/`‖`、`♥`、歌名、歌手）并记录每段含居中填充的绝对列区间，与视图一起缓存；点击和悬停一律经 `songInfoRenderer.elementAt(x)` 命中，不要在 `event_handler.go` 里另算宽度。
- **扁平进度条**：`[theme.progress] renderMode = "flat"`：已播放部分用 Secondary，播放头用 Accent，空槽用 `progressEmpty`，时间用 Secondary。此模式下 `internal/commands/netease.go` 启动时以主题 `muted`→`secondary` 调用 vendor `model.SetProgressColor`，启动页进度条同色而不再随机。
- **自定义表单页面**：`internal/ui/page_layout.go` 统一提供标题返回按钮、面包屑委托、输入框焦点/hover 样式、按钮本地化和文本光标定位；Search、Login 与 Last.fm API account 页面必须按其渲染结果记录命中坐标，并以 `pageSubmitButton` / `pageButton` 保持键盘和鼠标 active 状态一致。
- **封面图**：Kitty 封面图在应用背景预留的透明区域渲染，位于文本和单元格背景之下；popup 仅覆盖与其重叠的部分。背景排除矩形仅覆盖图像实际填充的前 `rows-1` 行，Kitty 未填充的末行仍由主题 `AppBackground` 绘制。独立页面若需清理封面，必须在 deferred `MenuToPage.BeforeEnterMenuHook` 中调用 `coverRenderer.ClearDisplayed()`，使删除发生在 loading 的 Main 帧之后、目标页面切换之前；Last.fm API account 入口遵循此规则。
- **网页登录**：macOS / Windows / Linux 三平台（非三平台 `login_webview.go` 桩）。Login 页「网页登录」入口在 WebView 检测不可用时自动隐藏：`webviewLoginAvailable()` 按 build tag 实现运行时检测——macOS WebKit 系统自带恒可用；Windows 经 `webviewloader.GetAvailableCoreWebView2BrowserVersionString("")` 检测 WebView2 Runtime；Linux 以 `webkitgtk.WebKitGTKVersion() == 0` 判定 WebKitGTK 库缺失；不可用时按钮不渲染，鼠标 hover/点击与键盘 right 导航同步跳过。检测结果在 `NewLoginPage` 缓存于 `LoginPage.webviewAvailable`。可用时弹出原生 WebView 窗口加载 `https://music.163.com/#/login`，通过原生 cookie API 每 ~1s 轮询检测 `MUSIC_U` cookie（HttpOnly，`document.cookie` 读不到）。页面层 `internal/ui/login_webview_page.go` 平台无关，仿 QR 登录页用 `tea.Tick` 轮询消费事件（`WebviewLoginEvent{CookieString, WindowClosed}`，缓冲 4 channel select-default 永不 close）；拿到 cookie 后复用 `utils/app.ParseCookieFromStr` → `neteaseutil.SetGlobalCookieJar` → `apputils.RefreshCookieJar()` 链路校验，失败则保持窗口继续轮询。共享常量在 `login_webview_consts.go`。三平台实现：
  - **macOS**：`internal/macdriver/webkit`（purego + objc 桥接，无 CGO）弹出 WKWebView，使用 `nonPersistentDataStore`（内存不落盘）；`WKHTTPCookieStore getAllCookies:`（macOS 27 改名，init 时经 `instancesRespondToSelector:` 兼容 `getAllCookiesWithCompletionHandler:`）轮询；block 回调运行在 WebKit 私有队列，只做数据 marshal 经 channel 回传。控制器 `internal/ui/login_webview_darwin.go` + 主线程派发 `login_webview_helper_darwin.go`（`MusicfoxWebviewLoginHelper`，仿 `internal/desktop_lyrics/helper_darwin.go` 的 performSelectorOnMainThread 模式）。窗口显示期间将全局 `SetActivationPolicy(Prohibited)`（`internal/runtime/runtime_darwin.go:26`）临时切换为 `Accessory` 并激活，关闭时恢复；所有退出路径（成功/用户关窗/取消）都恢复。
  - **Windows**：`internal/ui/login_webview_windows.go`，依赖 `github.com/13thgoutham/go-webview2` v1.0.25（wailsapp/go-webview2 v1.0.23 的修复 fork，零 CGO，`pkg/webview2` 生成 COM 绑定 + `webviewloader` 从内存加载 WebView2Loader.dll）。单 UI 线程模型：messageLoop goroutine（`LockOSThread` + CoInitializeEx）承载建窗（RegisterClassExW/CreateWindowExW）→ 环境/控制器创建（回调经消息泵派发）→ `SetTimer` WM_TIMER 轮询 `GetCookieManager().GetCookies("")`（空 uri 取全部 cookie 含 HttpOnly）→ 消息循环退出后 `cleanup()`（`controller.Close()` + 经 IUnknown vtable Release 槽释放 COM 对象）。Close() 幂等 = 置 closed + `PostMessage(WM_CLOSE)`。WebView2 Runtime：Win11 自带，Win10 绝大多数已装。**依赖替换原因**：wailsapp/go-webview2 v1.0.23 生成的 `ICoreWebView2AddScriptToExecuteOnDocumentCreatedCompletedHandler` / `ICoreWebView2CallDevToolsProtocolMethodCompletedHandler` / `ICoreWebView2ExecuteScriptCompletedHandler` 三个 handler 的 Invoke 回调以 `string`（16 字节）承载 LPCWSTR 结果，新版 Go（≥1.22）`compileCallback` 逐参数限制 ≤ uintptr（8 字节）导致包 init 时 panic；上游无修复版本（最新即 v1.0.23，PR #36/#37 未合并），故改用 13thgoutham fork v1.0.25（含 PR #36/#37 的生成器级修复）；`make vendor` 后无需额外打补丁。
  - **Linux**：`internal/webkitgtk`（purego 直调 WebKitGTK，零 CGO）运行时优先加载 6.0 栈（`libwebkitgtk-6.0.so.4` + `libgtk-4.so.1`），失败自动回退 4.1 栈（`libwebkit2gtk-4.1.so.0` + `libgtk-3.so.0`，覆盖 Debian 11/12、Ubuntu 22.04 等旧发行版）。6.0 使用 `WebKitNetworkSession` 与 `webkit_cookie_manager_get_all_cookies`（WebKitGTK 6.0 没有 `webkit_web_view_new_with_network_session`，会话要作为构造属性传入，所以这条路径目前总是不可用，以 debug 日志回退 4.1）；4.1/4.0 使用旧版 `WebContext` / `CookieManager` API，其中 4.1 链接 libsoup 3、4.0 链接 libsoup 2，并通过带 URI 的 `webkit_cookie_manager_get_cookies` 查询 Cookie。所有支持栈都不可用不 panic（`WebKitGTKVersion()` 返回 0，导出函数降级 no-op，controller 走 WindowClosed 回退，且 Login 页「网页登录」入口自动隐藏）。**符号缺失降级**：库可加载（RTLD_LAZY）但缺少对应 API 符号时，`registerSymbol`（`purego.Dlsym` 失败静默跳过而非 `RegisterLibFunc` panic）使对应函数指针保持 nil，栈可用性检查判定该栈不可用并继续回退，最终 `WebKitGTKVersion()` 返回 0 走表单登录降级，不影响主应用启动。版本差异分发：`gtk_window_new` 带参/无参、`gtk_window_set_child` vs `gtk_container_add`、`gtk_widget_show_all`、窗口关闭信号 "close-request"（GTK4）vs "delete-event"（GTK3，均经 `GSignalConnectCloseRequest` 统一连接，回调统一 `func(window, extra uintptr) int32` 返回 0 允许关闭）。`internal/ui/login_webview_linux.go`：GTK 线程模型——`runGTK` goroutine（`LockOSThread`）内 `gtk_init_check` → 创建对应 API 栈的 WebView → `gtk_window_new` → `g_timeout_add` 轮询 Cookie → `gtk_main` 阻塞；Close() 幂等 = `gtk_main_quit`（跨线程安全）。无 DISPLAY/WAYLAND_DISPLAY 时直接发 WindowClosed 降级二维码登录。每次轮询释放 GList/SoupCookie/GError，退出时 `g_object_unref` window/webView/session。

## 开发指南

### 添加新菜单

1. 创建 `internal/ui/menu_new_feature.go`，嵌入 `baseMenu`
2. 实现 `Menu` 接口
3. 注册到导航系统

### 添加新播放器引擎

1. 实现 `internal/player.Player` 接口
2. 在 `player.go:NewPlayerFromConfig()` 添加 case
3. 添加配置支持

### 修改快捷键

1. 在 `internal/keybindings/keybindings.go` 定义 `OperateType`
2. 在 `event_handler.go` 添加键映射
3. 在配置文件中添加自定义绑定

### 添加新渲染器

1. 实现 `Update()` 和 `View()` 方法
2. 在 `netease.go:Components()` 注册

### 更新依赖

使用 `make vendor` 更新 vendor 目录，该命令会依次执行 `go mod tidy`、`go mod vendor` 并复制 CGo 依赖的头文件。

```bash
make vendor
```

#### foxful-cli（内置源码）

foxful-cli 是本项目的核心 UI 框架，源码在 `third_party/foxful-cli/` 内置维护（独立 go.mod，模块路径仍为 `github.com/anhoder/foxful-cli`）。根 go.mod 以 `replace github.com/anhoder/foxful-cli => ./third_party/foxful-cli` 引用，require 保留上游基线版本，导入路径与上游 go-musicfox 一致，便于合并上游提交。

- `vendor/github.com/anhoder/foxful-cli/` 是 `make vendor` 生成的副本，禁止直接修改。改了 third_party 不执行 `make vendor` 时，默认的 vendor 模式构建仍用旧副本；`internal/forkcheck` 测试会在副本过期时失败。
- 开发：改 `third_party/foxful-cli` 后执行 `make vendor`（需要联网补全模块缓存；`modvendor` 装在 `$(go env GOPATH)/bin`，不在 PATH 时临时加入），再 `make test`。框架自身的测试用 `go test -mod=readonly github.com/anhoder/foxful-cli/model/... github.com/anhoder/foxful-cli/style/...`（`example/global_key` 需要 X11 开发头文件，不在测试范围内）。模块缓存完整时也可以 `GOFLAGS=-mod=readonly make build` 直接编译 third_party。
- 来源与基线记录在 `third_party/foxful-cli/UPSTREAM.md`。提交 `chore(foxful-cli): import pristine v1.1.4` 是未修改的上游原版，`git diff <该提交> -- third_party/foxful-cli` 即全部框架改动。
- 合并上游 go-musicfox 时如果 foxful-cli 的 require 版本变了，`TestFoxfulBaseVersion` 会失败：把「旧基线 → 新版本」的差异三方合并进 third_party，更新 UPSTREAM.md 的 Base，再 `make vendor`。vendor 冲突一律以重新生成为准。
- 上游仓库没有 LICENSE 文件；不要在 third_party 内添加声明上游许可的文件。
- 框架改动清单（相对上游 v1.1.4）：
  1. `style/theme.go`：`base.MenuItem` 应用 `highlights.menuItem` 前景色（上游只设背景，未选中菜单项显示终端默认色）；`StyleSet.Theme()` 访问器；Theme 与 StyleSet 新字段 `MenuSelectedArrow`、`MenuArrow`、`MenuIndex`、`StatusBarTimeIcon`（默认 `"\ue384 "`），Theme 新字段 `TabBorder`。
  2. `model/main.go`：`formatEntry` 与 `menuItemView` 使用 `ss.MenuSelectedArrow`，经 `styleMenuRow` 按显示宽度把一行切成标记、编号、标题三段分别着色（前缀总宽仍为 4+N+2，点击检测不变）；居中模式下的未选中行改用 `ss.MenuItem`（上游没有样式）；新增 `isNoColor`。
  3. `model/statusbar.go`：时钟前缀改用 `ss.StatusBarTimeIcon`。
  4. `model/tabs.go`：未聚焦标签的边框使用 `Theme().TabBorder`。
  5. `model/startup.go`：新增 `SetProgressColor(start, end)`。
  6. `model/startup_animation.go`：启动 logo 取 `Title` 前景色，未设置时回退 `util.GetPrimaryColor()`。
  7. `ansix` 包：按 Wc 宽度精确处理的 `Width`、`Cut`、`Fit`、`Shift`、`Truncate`、`Normalize`（x/ansi 的 `CutWc`、`TruncateWc` 在宽字符边界上不可靠）；`Fade`（只改 SGR 颜色、量化 32 级、每行自带样式）、`Restyle`（按列改样式）、`Mix`、`SameColor`。帧规则：交给渲染器的每一行按 Wc 计宽都等于窗口宽度。
  8. `motion` 包：`Level`（Full / Subtle / Off）、缓动、`Tween`、闭式临界阻尼 `Spring`、`StaggerDelay`，以及 `Scheduler`（同时只有一个帧定时器在途：快帧 16ms，慢帧 `SlowInterval` 50ms）。
  9. `model/app_motion.go`：`Options.Motion`（`Level`、`FrameInterval`、`GlideSteps`），`App.Animate` / `AnimateSlow` / `AnimateUntil`、`FrameTime` / `FrameSeq`、`SurfaceColors`、`TerminalVersion`、`CellAspect`；`Init` 请求终端前景色、背景色、XTVERSION 与 CSI 16t。没有调度器（测试）时弹窗、菜单、选中项都不做动画。
  10. `model/app_transition.go`：页面切换经背景交叉淡化（full 200ms / subtle 150ms）；`PageTransitionHandler` 让页面自己做过渡；`Snapshot` 离屏渲染（期间 `snapshotting` 为真，菜单过渡和选中滑动不更新）；`SuppressPageTransition`；离开启动页时第一个页面逐行进入。
  11. `model/main_layout.go`、`model/side_panel.go`、`model/tracklist.go`：Main 按 `ContentWidth()` 布局（P1 前后的金样帧逐字节一致，见 `main_golden_test.go`）；`SidePanel` 可以是右侧一栏（`Width`）或菜单下方的停靠条（`Height`），面板之下的组件成为全宽 footer（`FooterRows`）；`ListTracklist` 曲目表、`MenuHeader` 分节标题、`MenuItem.Meta` 右对齐时长；加载新菜单期间分节标题已显示新菜单名。
  12. `model/modal_motion.go`：弹窗与右键菜单淡入并上浮一行、关闭时淡出；通知从所在边滑入滑出。被移除的弹窗或通知留下只绘制不参与命中的「残影」，`HasPopup`、`TopModalBounds` 语义不变；超出窗口的图层先裁剪（lipgloss Compositor 会按图层并集扩大画布）。
  13. `model/main_transition.go`：菜单进出、翻页、切标签时旧行淡出漂移 2 格，新行逐行错开从 3 格外滑入（总长 ≤240ms，subtle 只淡入）；加载中、改变窗口尺寸、Main 上一帧不在屏幕上时不触发。
  14. `model/selection_glide.go`：曲目表选中项移动时色带交叉淡化 150ms，full 时一条左半格竖条按 `GlideSteps`（八分块 4 级、六分块 3 级、四分块 2 级）滑过去，停下后恢复主题的选中标记。
  15. `model/startup.go`：`StartupOptions.View` 钩子由宿主绘制启动页（自带背景，不再经过背景归一化）。
  16. `model/statusbar.go`、`style/theme.go`：状态栏左侧标签可由主题配置（`statusBar.label`，Warm Night 用 `■`）。
  17. `model/main_loading.go`：实现了 `BackgroundLoader`（`LoadsInBackground() == true`）的菜单，其 `BeforeEnterMenuHook` 在 UI goroutine 之外运行，结束后以 `menuLoadedMsg` 回到 `App.Update`（Main 不在屏幕上时也会进入菜单，但不切换页面）；加载期间界面照常绘制，分节标题已显示新菜单名；`BackMenu`、`BackToMenu`、切标签、`EnterMenu` 会取消进行中的加载，迟到的结果被丢弃。这类钩子不得接触界面（写终端、改页面状态）。`EnterMenu`、`DeferWithLoading`、`RefreshMenuWithLoading` 与 `BottomOutHook` 仍同步执行。

### 发布

版本号独立于上游，从 v1.0.0 开始；`types.UpstreamVersion` 记录所基于的 go-musicfox 版本（合并上游后更新），`--version` 一并显示。因为与原版共用数据目录：更新日志弹窗记在自己的键 `deck_changelog_seen`（`storage.ChangelogSeen`），从原版迁来的用户也会看到一次；`ExtInfo.StorageVersion` 与当前版本不同（无论新旧）就刷新通知图标等文件。`version.CompareVersion` 按数字比较各段（1.10 晚于 1.9）。

发布步骤：

1. 在 `CHANGELOG.md` 顶部写好本版一节：版本号、`=` 下划线、`YYYY-MM-DD` 日期，然后是中文条目，`<!-- en -->` 一行之后是对应的英文条目；与 go-musicfox 的记录之间以 `---` 隔开。同步程序内弹窗用的 `utils/filex/embed/changelog.md`（仅中文）。
2. 打 tag `vX.Y.Z` 并推送，`.github/workflows/release.yaml` 依次：
   - `draft`：以 `hack/release_notes.sh` 生成的说明创建草稿 Release（CHANGELOG 中文部分与中文下载表在前，英文部分与英文下载表在后）；
   - `linux-windows`：在 Ubuntu 22.04 容器里运行 `hack/release_container.sh`（Go、GoReleaser、llvm-mingw 都固定版本并校验 SHA-256），按 `.goreleaser.yaml` 编 Linux amd64/arm64 与 Windows amd64/arm64，打 tar.gz、zip、deb、rpm、Arch 包；`macos`：在 macOS runner 上按 `.goreleaser.darwin.yaml` 编两个架构的命令行版与 .app 版。两者都上传到同一个草稿（`use_existing_draft`）；
   - `try-*`：在 Ubuntu 22.04/24.04、Debian 12（另有 arm64，经 QEMU）、Fedora、openSUSE Tumbleweed、Arch 的容器里安装包并运行 `musicfox --version`；在 macOS 上运行两个架构的二进制与 .app 版，并用生成的 formula 安装测试；在 Windows 上运行 x64 版；
   - `checksums`：全部通过后上传 `SHA256SUMS`。
3. 检查草稿后手动发布（`gh release edit vX.Y.Z --draft=false`）。发布正式版会触发 `.github/workflows/homebrew.yaml`：`hack/brew_formula.sh` 生成 formula，经部署密钥（secret `HOMEBREW_TAP_DEPLOY_KEY`）推到 `Egdon/homebrew-tap`。

带 `-` 的 tag（如 `v1.0.0-rc.1`）是预发布：说明取去掉后缀的版本那一节，不更新 Homebrew。手动运行 release 工作流只构建、试装快照，不发布。

- **Linux**：`hack/pkg-config-static` 让 goflac 的 `pkg-config --libs flac` 返回静态链接的 libFLAC 与 libogg，二进制只依赖 glibc ≥ 2.34 与 `libasound.so.2`。各发行版的 libFLAC soname 不同（.so.8 / .12 / .14），上游的发布包动态链接 libFLAC.so.8，在新发行版上无法启动。
- **Windows**：只有 minimp3 需要 C 编译器；llvm-mingw（UCRT）编出的程序只依赖系统 DLL。
- **macOS**：未签名，只有链接器的 ad-hoc 签名。Homebrew formula 安装预编译的二进制，formula 下载的文件不带隔离属性，不会被 Gatekeeper 拦下；手动下载需要 `xattr -d com.apple.quarantine`。
- **本地预演**：`make release-snapshot` 用 podman（或 docker）在同样的容器里构建 Linux 与 Windows 快照，输出到 `dist/`；macOS 只能在 CI 上验证。

#### 打包 RPM（本机）

`hack/rpm.sh` 用系统 gcc 执行 `make build`（conda 等环境的编译器找不到系统的 FLAC、ALSA 头文件），再按 `deploy/rpm/musicfox-deck.spec` 把 `bin/musicfox` 打成本机用的 RPM，输出到 `dist/`。版本号取最近的 tag 加快照后缀（如 `1.0.0^20261020git11c1f7e`）。依赖由 rpmbuild 从二进制推导（本机的 `libFLAC.so.N`、`libasound.so.2`、glibc），只适合在打包的这台机器所用的发行版上安装；分发用 Release 里的 rpm。WebKitGTK 只在网页登录时动态加载，列为推荐依赖。该脚本只用于 Linux。

### 跨平台构建兼容性

**修改 `Makefile` 或 `hack/` 目录下的构建脚本时，必须确保 Windows 系统兼容。**

#### 原因
go-musicfox 支持 macOS/Linux/Windows 三平台，构建脚本需要在不同环境下正确执行。

#### 检查清单
- [ ] 新增的 target 是否包含 Unix 特有命令（`which`、`cp`、`mkdir -p`、`chmod`、`tar`、`awk` 等）
- [ ] 是否使用 `$(OS)` / `Windows_NT` 条件分支提供 Windows 替代逻辑
- [ ] Shell 脚本（`.sh`）是否有对应的 PowerShell（`.ps1`）实现
- [ ] 路径分隔符是否兼容（使用 `$(PACKAGE_ROOT)` 而非 `` `pwd` ``）
- [ ] 重定向和设备文件是否有 Windows 替代（`nul` 对应 `/dev/null`，`where` 对应 `which`）

#### Windows 兼容模式参考

```makefile
# 使用 OS 条件分支
ifeq ($(OS),Windows_NT)
    # Windows 逻辑
else
    # Unix 逻辑
endif
```

| Unix 写法 | Windows 替代 |
|-----------|-------------|
| `which <cmd>` | `where <cmd>` |
| `/dev/null` | `nul` |
| `` `pwd` `` | `$(PACKAGE_ROOT)`（Make 变量） |
| `<cmd> >/dev/null 2>&1` | `<cmd> >nul 2>&1` |
| `<cmd> || { cmd2; }` | `<cmd> || ( cmd2 )` |
| `hack/*.sh` | `hack/*.ps1`（PowerShell 实现） |

**例外**：纯交叉编译/CI 内部工具（如 `hack/init_linux_env.sh`、`hack/init_windows_env.sh` 等 Docker 内部脚本，以及发布用的 `hack/release_container.sh`、`hack/pkg-config-static`、`hack/release_notes.sh`、`hack/brew_formula.sh` 和 `make release-snapshot`），仅在 Linux 容器或 CI 中执行，无需适配 Windows。

## 文档维护准则

### 重要准则：修改代码后需维护 AGENTS.md

**所有贡献者在修改代码后，必须检查并更新 AGENTS.md 文档，防止文档腐化。**

#### 何时需要更新文档
- 添加、删除或重命名核心文件或目录
- 新增功能模块或组件
- 修改项目结构或架构
- 更改 API 接口或配置格式
- 添加新的播放器引擎、菜单类型、渲染模式
- 修改关键路径或重要流程
- **快捷键变更**：新增、删除或修改快捷键后，必须同步更新 README.md 与 README.en.md 中的快捷键说明部分（help 菜单是动态从 keybindings 配置生成的，无需手动修改）
- **README 双语**：README.md（简体中文，主版本）与 README.en.md（英文）逐节对应，修改其一时必须同步另一份。文档使用书面说明文的写法，避免口语化表述

#### 更新检查清单
- [ ] 目录结构是否准确反映当前项目结构
- [ ] 核心文件路径是否正确
- [ ] 接口定义是否与代码一致
- [ ] 新增功能的说明是否完整
- [ ] 开发指南是否需要补充
- [ ] 关键文件路径表格是否需要更新

#### 更新优先级
| 变更类型 | 优先级 | 说明 |
|---------|--------|------|
| 架构变更 | **高** | 必须立即更新 |
| 新增核心模块 | **高** | 必须添加说明 |
| 文件路径变更 | **中** | 及时更新路径表 |
| 细微修复 | **低** | 可批量更新 |

#### 维护建议
- 保持文档与代码同步，避免技术债务积累
- 使用一致的术语和格式
- 添加代码示例时确保与实际代码匹配
- 定期（如每月）审查文档完整性
- PR 审查时应包含文档检查

#### 文档腐化警告信号
- 文件路径与实际不符
- 过时的 API 接口描述
- 已删除功能仍出现在文档中
- 章节结构混乱或重复
- 与 README/CHANGELOG 存在矛盾

违反此准则可能导致：新贡献者无法理解项目结构、开发效率降低、文档失去参考价值、维护成本增加。

## Git 提交规范

### 重要准则：Git Commit Message 必须遵循 Conventional Commits 规范

**所有贡献者在提交代码时，必须遵循 [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) 规范。**

#### 提交格式

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Type 类型

| Type | 说明 |
|------|------|
| `feat` | 新功能开发 |
| `fix` | Bug 修复 |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响功能） |
| `refactor` | 代码重构（既不修复bug也不添加功能） |
| `perf` | 性能优化 |
| `test` | 测试相关 |
| `chore` | 构建工具、辅助工具、配置变更 |
| `revert` | 回滚提交 |

#### 示例

```
feat(player): 添加 MPV 播放引擎支持

- 支持多种音频格式
- 实现播放进度控制
- 优化内存使用

Closes #123
```

#### 规范要求

1. **必须使用英文**：提交信息、描述均使用英文
2. **动词开头**：描述部分以动词开头，使用现在时态
3. **长度限制**：标题不超过 50 字符
4. **Body 可选**：复杂变更可添加详细说明，每行不超过 72 字符
5. **引用 Issues**：在 footer 中使用 `Closes #xxx` 关联 Issue
