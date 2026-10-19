// Table example — demonstrates data table with navigation using the App framework.
//
// This example shows multiple table types (products, users, transactions) accessible
// through a menu. Each table demonstrates keyboard navigation, row selection, and
// status display.
//
// Navigation:
//
//	↑↓/jk         — navigate menu or table rows
//	Enter         — select table type
//	PgUp/PgDn     — page up/down in table
//	g/Home        — jump to first row
//	G/End         — jump to last row
//	b/esc         — back to menu
//	q/Ctrl+C      — quit
package main

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/anhoder/foxful-cli/model"
	"github.com/anhoder/foxful-cli/style"
)

var mainMenu = NewTableMenu()

// Sample data for different table types
var (
	productRows = [][]string{
		{"001", "Mechanical Keyboard", "Peripherals", "$129.99", "47"},
		{"002", "27\" 4K Monitor", "Displays", "$399.00", "12"},
		{"003", "USB-C Hub 7-in-1", "Accessories", "$49.95", "134"},
		{"004", "Ergonomic Mouse", "Peripherals", "$79.00", "88"},
		{"005", "Laptop Stand", "Accessories", "$35.00", "202"},
		{"006", "Webcam 1080p", "Peripherals", "$89.99", "55"},
		{"007", "SSD 1TB NVMe", "Storage", "$109.00", "67"},
		{"008", "RAM 32GB DDR5", "Memory", "$149.00", "30"},
		{"009", "GPU RTX 4070", "Graphics", "$599.00", "8"},
		{"010", "CPU Cooler 240mm", "Cooling", "$69.99", "41"},
		{"011", "ATX Power Supply", "Components", "$89.00", "25"},
		{"012", "Cable Management Kit", "Accessories", "$18.00", "315"},
		{"013", "Desk Mat XL", "Accessories", "$39.99", "97"},
		{"014", "Monitor Arm", "Displays", "$55.00", "73"},
		{"015", "Headset USB", "Audio", "$59.00", "44"},
	}

	userRows = [][]string{
		{"1001", "Alice Chen", "alice@example.com", "Engineering", "Active"},
		{"1002", "Bob Smith", "bob@example.com", "Sales", "Active"},
		{"1003", "Carol White", "carol@example.com", "Marketing", "Inactive"},
		{"1004", "Dave Johnson", "dave@example.com", "Engineering", "Active"},
		{"1005", "Eve Brown", "eve@example.com", "HR", "Active"},
		{"1006", "Frank Lee", "frank@example.com", "Finance", "Pending"},
		{"1007", "Grace Taylor", "grace@example.com", "Engineering", "Active"},
		{"1008", "Henry Wilson", "henry@example.com", "Operations", "Active"},
		{"1009", "Iris Martinez", "iris@example.com", "Sales", "Active"},
		{"1010", "Jack Davis", "jack@example.com", "Engineering", "Inactive"},
		{"1011", "Kate Anderson", "kate@example.com", "Marketing", "Active"},
		{"1012", "Leo Garcia", "leo@example.com", "Support", "Active"},
		{"1013", "Maya Rodriguez", "maya@example.com", "Engineering", "Active"},
		{"1014", "Nick Thompson", "nick@example.com", "Finance", "Pending"},
		{"1015", "Olivia Moore", "olivia@example.com", "HR", "Active"},
	}

	transactionRows = [][]string{
		{"TXN-9001", "2026-07-20", "Payment", "$1,250.00", "Completed"},
		{"TXN-9002", "2026-07-21", "Refund", "$49.99", "Completed"},
		{"TXN-9003", "2026-07-21", "Payment", "$899.00", "Pending"},
		{"TXN-9004", "2026-07-22", "Payment", "$2,340.50", "Completed"},
		{"TXN-9005", "2026-07-22", "Adjustment", "$25.00", "Completed"},
		{"TXN-9006", "2026-07-23", "Payment", "$175.00", "Failed"},
		{"TXN-9007", "2026-07-23", "Payment", "$450.00", "Completed"},
		{"TXN-9008", "2026-07-23", "Refund", "$89.99", "Pending"},
		{"TXN-9009", "2026-07-24", "Payment", "$3,200.00", "Completed"},
		{"TXN-9010", "2026-07-24", "Payment", "$67.50", "Completed"},
		{"TXN-9011", "2026-07-24", "Payment", "$540.00", "Processing"},
		{"TXN-9012", "2026-07-24", "Chargeback", "$125.00", "Under Review"},
	}
)

// ── TableMenu ────────────────────────────────────────────────────────

type TableMenu struct {
	model.DefaultMenu
	menus []model.MenuItem
}

func NewTableMenu() *TableMenu {
	return &TableMenu{
		menus: []model.MenuItem{
			{Title: "Product Inventory", Subtitle: "Browse 15 products with pricing and stock"},
			{Title: "User Directory", Subtitle: "View 15 users with departments and status"},
			{Title: "Transaction History", Subtitle: "Review 12 recent transactions"},
		},
	}
}

func (m *TableMenu) GetMenuKey() string {
	return "table_menu"
}

func (m *TableMenu) MenuViews() []model.MenuItem {
	return m.menus
}

func (m *TableMenu) IsSearchable() bool {
	return true
}

func (m *TableMenu) SubMenu(_ *model.App, _ int) model.Menu {
	return nil
}

func (m *TableMenu) Action(app *model.App, index int) (model.Page, tea.Cmd) {
	// Navigate to table page based on selection
	switch index {
	case 0:
		return NewTablePage(app, "products"), nil
	case 1:
		return NewTablePage(app, "users"), nil
	case 2:
		return NewTablePage(app, "transactions"), nil
	}
	return nil, nil
}

// ── TablePage ────────────────────────────────────────────────────────

type TablePage struct {
	app       *model.App
	tableType string
	table     *model.Table
	width     int
	height    int
}

func NewTablePage(app *model.App, tableType string) *TablePage {
	p := &TablePage{
		app:       app,
		tableType: tableType,
	}

	// Create table based on type
	switch tableType {
	case "products":
		p.table = p.createProductTable()
	case "users":
		p.table = p.createUserTable()
	case "transactions":
		p.table = p.createTransactionTable()
	}

	p.table.Focus()
	return p
}

func (p *TablePage) createProductTable() *model.Table {
	columns := []model.Column{
		{Title: "ID", Width: 4},
		{Title: "Product", Width: 22},
		{Title: "Category", Width: 12},
		{Title: "Price", Width: 9},
		{Title: "Stock", Width: 0}, // flex width
	}
	return model.NewTable(columns, productRows)
}

func (p *TablePage) createUserTable() *model.Table {
	columns := []model.Column{
		{Title: "ID", Width: 5},
		{Title: "Name", Width: 18},
		{Title: "Email", Width: 22},
		{Title: "Department", Width: 12},
		{Title: "Status", Width: 0}, // flex width
	}
	return model.NewTable(columns, userRows)
}

func (p *TablePage) createTransactionTable() *model.Table {
	columns := []model.Column{
		{Title: "Transaction ID", Width: 10},
		{Title: "Date", Width: 12},
		{Title: "Type", Width: 12},
		{Title: "Amount", Width: 11},
		{Title: "Status", Width: 0}, // flex width
	}
	return model.NewTable(columns, transactionRows)
}

func (p *TablePage) Type() model.PageType {
	return model.PageType("table_" + p.tableType)
}

func (p *TablePage) Msg() tea.Msg {
	return nil
}

func (p *TablePage) IgnoreQuitKeyMsg(_ tea.KeyMsg) bool {
	return false
}

func (p *TablePage) Update(msg tea.Msg, a *model.App) (model.Page, tea.Cmd) {
	p.app = a

	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := msg.String()
		switch key {
		case "esc", "b", "B":
			return a.Main(), nil
		}

		// Forward to table widget
		cmd := p.table.Update(msg)
		return p, cmd

	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		// Reserve space for header, status, controls
		tableH := p.height - 10
		if tableH < 5 {
			tableH = 5
		}
		p.table.SetSize(p.width-4, tableH)
		return p, nil
	}

	return p, nil
}

func (p *TablePage) View(a *model.App) string {
	w, h := a.WindowWidth(), a.WindowHeight()
	if w <= 0 || h <= 0 {
		return ""
	}

	// Set widget size based on current window dimensions
	tableH := h - 10
	if tableH < 5 {
		tableH = 5
	}
	p.table.SetSize(w-4, tableH)

	ss := style.CurrentStyleSet()

	// Header based on table type
	var header, subtitle string
	switch p.tableType {
	case "products":
		header = "📦 Product Inventory"
		subtitle = "Browse available products with pricing and stock information"
	case "users":
		header = "👥 User Directory"
		subtitle = "View all registered users and their current status"
	case "transactions":
		header = "💳 Transaction History"
		subtitle = "Review recent payment and financial transactions"
	}

	title := ss.Title.Render(header)
	sub := ss.Muted.Render(subtitle)

	// Table wrapped in a border box
	tableContent := p.table.View()
	tableBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ss.Border.GetForeground()).
		Padding(0, 1).
		Width(p.width - 4).
		Render(tableContent)

	// Status bar with selected row info
	status := p.renderStatus(ss)

	// Controls
	controls := ss.Muted.Render("↑↓/jk: navigate • PgUp/PgDn: page • g/G: first/last • b/esc: back")

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		title,
		sub,
		"",
		tableBox,
		status,
		"",
		controls,
	)

	return content
}

func (p *TablePage) renderStatus(ss style.StyleSet) string {
	selected := p.table.SelectedRow()
	total := len(p.table.Rows())

	if selected < 0 || selected >= total {
		return ss.Muted.Render(" No rows ")
	}

	row := p.table.Rows()[selected]
	var statusText string

	switch p.tableType {
	case "products":
		// ID, Product, Category, Price, Stock
		statusText = fmt.Sprintf(" Row %d of %d  •  %s  •  %s  •  %s in stock ",
			selected+1, total, row[1], row[3], row[4])
	case "users":
		// ID, Name, Email, Department, Status
		statusText = fmt.Sprintf(" Row %d of %d  •  %s  •  %s  •  %s ",
			selected+1, total, row[1], row[3], row[4])
	case "transactions":
		// Transaction ID, Date, Type, Amount, Status
		statusText = fmt.Sprintf(" Row %d of %d  •  %s  •  %s  •  %s  •  %s ",
			selected+1, total, row[0], row[2], row[3], row[4])
	}

	return ss.SelectedItem.Render(statusText)
}

// ── main ────────────────────────────────────────────────────────────

func main() {
	ops := model.DefaultOptions()
	ops.EnableStartup = false
	ops.WhetherDisplayTitle = true
	ops.AppName = "Table Demo"
	ops.Ticker = model.DefaultTicker(500 * time.Millisecond)

	app := model.NewApp(ops)
	app.With(model.WithMainMenu(mainMenu, nil))

	fmt.Println(app.Run())
}
