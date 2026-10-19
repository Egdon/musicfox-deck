---
name: foxful-tui-development
description: 在 foxful-cli 中开发或调试涉及 Bubble Tea 消息路由、终端 Cell 几何、键鼠交互、模态层、主题或可见组件行为的改动时使用；普通 Go 工具函数和纯数据改动走常规开发流程。
compatibility: opencode
metadata:
  project: foxful-cli
  language: go
---

# Foxful TUI 开发与调试

把每次 TUI 改动当作一条 **事件—渲染闭环**：

```text
tea.Msg → 路由 → 状态变更 → tea.Cmd/重绘 → View → 几何捕获 → 下一次输入
```

只有当真实输入走完这条闭环并出现可观察结果时，改动才算被理解和验证。

## 进入正确分支

- 首次进入仓库、构建基线未知或不清楚模块边界：先读取 [ONBOARDING.md](ONBOARDING.md)。完成标准是能指出状态所有者、输入入口、渲染出口和最近的测试或 example。
- 修复 Bug、性能回归、偶发交互或终端差异：读取 [DEBUGGING.md](DEBUGGING.md)。
- 任何可见行为、输入、尺寸、主题或组件契约变更：读取 [TESTING.md](TESTING.md)。

如果当前 Agent 能看到以下全局 Skills，则按需联动；仓库内参考仍是缺少这些 Skills 时的完整回退：

- `karpathy-guidelines`：编码或重构前，用于约束改动范围和成功标准；
- `verification-planning`：非平凡行为变更前，用于建立证据路径；
- `diagnosing-bugs`：困难 Bug、性能问题或无法稳定复现的问题。

## 先定位闭环

在编辑前明确四个事实：

1. 哪个对象拥有要改变的状态；
2. 哪个具体 `tea.Msg` 到达哪层 `Update`/handler；
3. 哪个 `View`/render 函数把状态变成终端 Cell；
4. 哪个现有测试或 example 能到达同一条路径。

只读取对应接口、实现、最近测试和最近 example。精确路由顺序以当前源码为准，不从 Skill 中推断。

**完成标准：** 用一行写出 `消息 → 路由 → 状态 → 渲染/动作`，并指出一个可能被测试捕获的回归。

## 稳定边界

- `App` 拥有活动 `Page`、模态栈、通知、终端尺寸和最终合成。
- `Page` 是公开的全屏扩展点；状态在 `Update` 中变化，`View` 从当前状态渲染。
- `Main` 拥有菜单层级、分页、搜索、Tabs、内置控制器、菜单几何和 Components。
- `Modal` 是内部协议；自定义模态内容通过 `Popup`、Markdown Popup、Form 和 Actions 接入。
- `Theme` 是配置输入，`StyleSet` 是解析后的渲染状态；新增视觉状态应进入既有主题体系。
- Lip Gloss 与 `layout` 使用终端 Cell；尺寸依据可视宽度和渲染高度，而不是字节数或 rune 数。
- 鼠标命中区域来自渲染后的几何；先 render/compose，再保存 bounds，最后处理后续鼠标消息。
- 零值配置保持既有行为，除非需求明确改变默认契约。

## 开发步骤

### 1. 写可观察契约

使用以下形状：

```text
给定 <Options + 终端尺寸 + 初始状态>，当 <具体 tea.Msg> 到达时，
状态应发生 <转换>，并在 <渲染 Cell / Page / Cmd / Action> 上可观察。
```

列出真正相关的边界：空集合、零值、窄/矮终端、ANSI、CJK/Emoji、禁用项、不可见项、滚动或缩放边界。

### 2. 复用现有接缝

- 配置进入 `Options`、聚焦的 option struct 或既有 Spec；
- 输入使用 Bubble Tea v2 的真实消息类型；
- 渲染复用 `Theme`/`StyleSet`、`lipgloss.Width/Height` 和 `layout`；
- 异步结果通过 `tea.Cmd` 或消息回到更新闭环；
- 用户可见能力扩展最近的 `example/<feature>`，避免建立第二套示例约定。

修改导出符号前查全引用；同一次改动迁移所有调用方。

### 3. 走通整条闭环

确认消息被有意消费或转发、状态由正确对象更新、需要时返回重绘/动作命令、`View` 呈现新状态，并保持 selection、focus、hover、scroll 和 bounds 一致。

**完成标准：** 直接测试或可运行 example 通过真实消息路径观察到需求行为。

## 调试步骤

按 [DEBUGGING.md](DEBUGGING.md) 建立可失败的反馈闭环，再从当前源码追踪消息在哪一跳偏离预期。一次只验证一个可证伪假设；修复后重跑最初场景，而不只运行缩小后的测试。

## 证据与交付

按 [TESTING.md](TESTING.md) 选择最低但忠实的测试接缝。可见 TUI 改动必须运行最近的 example；键鼠、滚动、拖拽或尺寸问题必须发送对应的真实消息或 PTY 输入。

交付前满足所有适用条件：

- 已命名状态所有者和事件—渲染闭环；
- 已执行真实的键盘、鼠标、滚轮或 WindowSize 输入；
- 已按终端 Cell 测量受影响尺寸；
- 行为测试覆盖新契约或旧故障；
- 最近 example 已真实运行并观察可见结果；
- 受影响包测试通过；
- 全仓测试结果或精确环境阻塞已记录；
- 临时日志、Harness 和调试状态已清理。
