# 测试、Harness 与 PTY 验收

可见行为、输入、尺寸、主题或组件契约变化时读取本文件。

若当前 Agent 可加载 `verification-planning`，先用它选择证据路径；若可加载 `karpathy-guidelines`，用它约束假设、改动范围和成功标准。本文件始终负责 foxful-cli 的具体测试接缝与命令。

## Go 测试接缝

使用真实 Bubble Tea v2 消息：

```go
tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
tea.MouseClickMsg(tea.Mouse{X: x, Y: y, Button: tea.MouseLeft})
tea.MouseWheelMsg(tea.Mouse{X: x, Y: y, Button: tea.MouseWheelDown})
tea.WindowSizeMsg{Width: 80, Height: 24}
```

渲染组件的测试顺序：

1. 使用确定的 `style.NewStyleSet(...)`；
2. render；
3. 通过 `lipgloss.Width/Height` 测量；
4. 组件需要命中测试时保存渲染生成的 bounds；
5. 发送真实输入；
6. 重渲染并断言可见输出与状态不变量。

仅在断言纯文本时 strip ANSI。颜色、背景归属、样式属性、Cell 宽度或层合成测试保留 ANSI。

## App 级临时 Harness

模板位于 [templates/tui_harness_test.go](templates/tui_harness_test.go)。复制到 `model/` 后，它提供：

- 固定尺寸的 `App + Page` 驱动器；
- Key、Text、Click、Wheel 消息 helper；
- `tea.View.Content`、纯文本视图和 Cell 尺寸；
- 默认不执行 Cmd 的 `Step`；
- 显式、最大步数和单命令超时受限的 `Drain`，支持直接 Cmd 与公开 `tea.BatchMsg`。

使用方式：

```text
cp .opencode/skills/foxful-tui-development/templates/tui_harness_test.go model/tui_harness_skill_test.go
go test ./model -run '^TestSkillTUIHarnessRoutesAppMessages$'
```

在模板上改出最小复现。修复后将有长期价值的断言转成正式测试，再删除临时文件。`Drain` 不自动执行 ticker、无限命令链或 Bubble Tea 内部不可见的 sequence 语义；遇到这些路径时由测试显式控制消息。

## 聚焦与包级命令

迭代时运行最窄的忠实测试：

```text
go test ./model -run '^TestRelevantBehavior$'
go test ./style -run '^TestRelevantTheme$'
go test ./example/menu
```

功能稳定后运行受影响包，再尝试全仓：

```text
go test ./model ./style ./layout
go test ./...
```

### 条件化探测 native 阻塞

不预设全仓测试必然失败。只有当前 `go test ./...` 实际失败后才分类：

1. 确认失败是否仅来自 `example/global_key` 或 `github.com/robotn/gohook`；
2. 记录缺失 header、CGO 或平台符号的原始错误；
3. 显式重跑所有受影响包；
4. 报告“受影响包通过、全仓被该环境阻塞”，而不是报告全仓绿色；
5. 若失败落在改动路径，按真实回归处理。

## 参数化 PTY Smoke

脚本位于 [scripts/tui-smoke.exp](scripts/tui-smoke.exp)。它使用 Expect 创建真实 PTY；缺少 Expect 时返回明确能力错误并提示改用 Go Harness。

示例：

```text
.opencode/skills/foxful-tui-development/scripts/tui-smoke.exp \
  --cols 120 --rows 30 --timeout 20 --ready 'Title 1' \
  --click right:10:10 \
  --key j --key j --key j \
  --expect 'Play Next' --expect '█' \
  -- go run ./example/menu
```

语义动作按参数出现顺序执行：

- `--key NAME`：方向键、Enter、Esc、Tab、Home、End、PgUp、PgDown、Ctrl-C 或单字符；
- `--text TEXT`：输入原始文本；
- `--click BUTTON:X:Y`：left/middle/right，协议坐标为 1-based；
- `--wheel DIRECTION:X:Y`：up/down/left/right；
- `--wait MS`：显式等待。

最后一个非 wait 动作前清空旧输出；动作后等待稳定窗口，仅对该窗口的 ANSI 归一化输出执行 `--expect`/`--reject` 正则断言。原始日志和归一化日志默认写入系统临时目录，脚本会打印路径；`--log-dir` 可覆盖。

PTY 证据必须包含终端尺寸、启动命令、动作序列、断言模式和日志路径。完整画面快照不是默认契约，避免动画和终端实现造成脆弱 diff。

## 验收完成标准

- 行为测试能在错误实现上失败；
- 用户输入走真实消息或 PTY 协议；
- 几何按终端 Cell 测量；
- 可见改动已运行最近 example；
- 受影响包通过；
- 全仓结果或当前环境阻塞已准确记录；
- 临时 Harness 与日志已处理。
