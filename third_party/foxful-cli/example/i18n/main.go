package main

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/anhoder/foxful-cli/model"
)

type localeMenu struct {
	model.DefaultMenu
	items []model.MenuItem
}

func newLocaleMenu() *localeMenu {
	return &localeMenu{items: []model.MenuItem{
		{Title: "English Demo", Subtitle: "en"},
		{Title: "简体中文演示", Subtitle: "zh-CN"},
		{Title: "日本語デモ", Subtitle: "ja"},
	}}
}

func (m *localeMenu) GetMenuKey() string          { return "i18n_locales" }
func (m *localeMenu) MenuViews() []model.MenuItem { return m.items }
func (m *localeMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}
func (m *localeMenu) IsSearchable() bool { return false }
func (m *localeMenu) HelpHints() []model.HelpHint {
	return []model.HelpHint{
		{Key: "up/down", Desc: "choose locale"},
		{Key: "enter", Desc: "apply locale"},
		{Key: "q", Desc: "quit"},
	}
}

func (m *localeMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	locales := []string{"en", "zh-CN", "ja"}
	if index >= 0 && index < len(locales) {
		model.SetLocale(locales[index])
	}
	return nil, app.RerenderCmd(true)
}

type translationComponent struct{}

func (c *translationComponent) Update(_ tea.Msg, _ *model.App) {}

func (c *translationComponent) View(app *model.App, _ *model.Main) (string, int) {
	styles := app.StyleSet()
	locale := model.DefaultCatalog().Locale()
	lines := []string{
		"",
		styles.Title.Render("Current locale: " + locale),
		styles.Border.Render(strings.Join([]string{
			styles.MenuTitle.Render("Live translations"),
			"",
			model.T(model.MsgLoading),
			model.T(model.MsgNoData),
			model.T(model.MsgEmptyDirectory),
			model.T(model.MsgFieldRequired),
			model.Tf(model.MsgReadError, "config.yaml"),
			"",
			"[" + model.T(model.MsgConfirm) + "]  [" + model.T(model.MsgCancel) + "]",
			model.T(model.MsgHintNavigate) + "  |  " + model.T(model.MsgHintQuit),
		}, "\n")),
	}
	return strings.Join(lines, "\n"), len(lines) + 10
}

func registerLocales() {
	catalog := model.DefaultCatalog()
	catalog.Register("zh-CN", map[model.MessageID]string{
		model.MsgLoading:        "正在加载...",
		model.MsgNoData:         "暂无数据",
		model.MsgEmptyDirectory: "（空目录）",
		model.MsgFieldRequired:  "此字段为必填项",
		model.MsgReadError:      "读取错误：%s",
		model.MsgConfirm:        "确认",
		model.MsgCancel:         "取消",
		model.MsgHintNavigate:   "导航",
		model.MsgHintQuit:       "退出",
	})
	catalog.Register("ja", map[model.MessageID]string{
		model.MsgLoading:        "読み込み中...",
		model.MsgNoData:         "データがありません",
		model.MsgEmptyDirectory: "（空のディレクトリ）",
		model.MsgFieldRequired:  "この項目は必須です",
		model.MsgReadError:      "読み込みエラー: %s",
		model.MsgConfirm:        "確認",
		model.MsgCancel:         "キャンセル",
		model.MsgHintNavigate:   "移動",
		model.MsgHintQuit:       "終了",
	})
}

func main() {
	registerLocales()
	model.SetLocale("en")

	options := model.DefaultOptions()
	options.AppName = "i18n Demo"
	options.MainMenuTitle = &model.MenuItem{Title: "Select a locale"}
	options.MainMenu = newLocaleMenu()
	options.Components = []model.Component{&translationComponent{}}
	options.BottomHeight = 14

	app := model.NewApp(options)
	fmt.Println(app.Run())
}
