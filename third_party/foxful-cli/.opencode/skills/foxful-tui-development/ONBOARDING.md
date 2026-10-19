# 首次接入与项目地图

仅在首次进入仓库、构建基线未知或需要重新建立架构心智模型时读取本文件。

## 环境基线

1. 读取 `go.mod`，以其中的 Go 版本和依赖声明为准；运行 `go version` 确认本机工具链。
2. 仓库提交了 `vendor/`。先使用仓库当前依赖闭环，不在无关任务中升级或重新整理依赖。
3. 运行核心包基线：

```text
go test ./model ./style ./layout
```

4. 运行与任务最近的 example 包测试，例如：

```text
go test ./example/menu
```

5. 全仓测试和 native 依赖的条件化判断按 [TESTING.md](TESTING.md) 执行。

**完成标准：** 能区分代码失败、工具链不匹配、vendor/native 依赖失败，并获得至少一条可重复的绿色核心包命令。

## 稳定项目地图

| 关注点 | 入口文件 |
|---|---|
| App 生命周期、事件优先级、Page 切换、模态与通知合成 | `model/app.go`, `model/page.go` |
| 主菜单、分页、搜索、Tabs、控制器、指针状态 | `model/main.go`, `model/menu.go`, `model/tabs.go`, `model/controller.go` |
| Popup、ContextMenu、滚动、缩放、命中测试 | `model/modal.go`, `model/popup.go`, `model/context_menu.go` |
| Markdown 与 Form 模态内容 | `model/markdown.go`, `model/form.go` |
| Table、Tree、FilePicker 等复用组件 | `model/table.go`, `model/tree.go`, `model/filepicker.go` |
| Main 下方附加区域与状态栏 | `model/component.go`, `model/statusbar.go`, `model/progress.go` |
| 主题输入、可访问性和解析后样式 | `style/theme.go`, `style/accessibility.go` |
| 终端 Cell 布局和层合成 | `layout/layout.go` |
| 真实运行入口 | `example/*/main.go` |

该表只记录稳定入口。字段、调用顺序和消息优先级必须从当前源码确认。

## 建立事件—渲染心智模型

对目标功能完成一次定向追踪：

1. 从具体 `tea.Msg` 类型找到最先接收它的 `Update` 或 handler；
2. 沿当前代码确认通知、模态、Page、Main、Component 中谁消费或继续转发；
3. 找到状态实际写入点；
4. 找到负责渲染该状态的 `View`/render；
5. 若涉及鼠标，找到 render 后将相对几何转成绝对 bounds 的位置；
6. 找到同目录测试和最近 example。

避免把 `Modal` 当作公开扩展协议。新的全屏能力优先实现 `Page`，新的模态内容优先使用现有 `Popup` 能力，新的视觉状态进入 `Theme`/`StyleSet`。

## 首次接入完成标准

开始实现前应能回答：

- 当前任务属于哪个状态所有者？
- 输入由哪种真实 Bubble Tea v2 消息触发？
- 状态在哪个更新函数改变？
- 可见结果由哪个渲染函数产生？
- 哪个测试接缝最忠实？
- 哪个 example 可做 PTY 验收？
