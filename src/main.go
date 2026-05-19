package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	text "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	purple = lipgloss.Color("#9d4edd")
	cyan   = lipgloss.Color("#4cc9f0")
	green  = lipgloss.Color("#4acc96")
	gray   = lipgloss.Color("#4a4e69")
	white  = lipgloss.Color("#f8f9fa")
	red    = lipgloss.Color("#ff4d6d")

	chatBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(purple).
			Padding(1, 2).
			Width(76).
			Height(18)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purple).
			Padding(0, 3).
			MarginBottom(1)

	statusStyle = lipgloss.NewStyle().Foreground(gray).Italic(true)
	myNickStyle = lipgloss.NewStyle().Foreground(cyan).Bold(true)
	peerStyle   = lipgloss.NewStyle().Foreground(green).Bold(true)
)

type Config struct {
	Nick string `json:"nick"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

type repoMsg string
type connectedMsg struct {
	conn net.Conn
	err  error
}

type model struct {
	nick      string
	messages  []string
	conn      net.Conn
	input     textinput.Model
	viewport  viewport.Model
	isHost    bool
	puerto    string
	ip        string
	connected bool
	dbPath    string
	peerNick  string
}

var configFile string

func init() {
	home, _ := os.UserHomeDir()
	configFile = filepath.Join(home, ".shchat.json")
}

func readNetMessages(conn net.Conn) text.Cmd {
	return func() text.Msg {
		scanner := bufio.NewScanner(conn)
		if scanner.Scan() {
			return repoMsg(scanner.Text())
		}
		return repoMsg("DISCONNECT_EVENT")
	}
}

func main() {
	fmt.Print("\033[H\033[2J")
	config := loadConfig()

	fmt.Println(titleStyle.Render(" GOChat Setup "))
	nick := readInput("Nick", config.Nick)
	fmt.Println("\n 1) Abrir sala (Host)\n 2) Conectar a sala (Cliente)")
	opcion := readInput("Opción", "1")

	m := model{
		nick:     nick,
		isHost:   opcion == "1",
		peerNick: "Usuario",
	}

	home, _ := os.UserHomeDir()
	m.dbPath = filepath.Join(home, ".gochat_history.json")

	if m.isHost {
		m.puerto = readInput("Puerto", config.Port)
		config.Nick = nick; config.Port = m.puerto; saveConfig(config)
		m.messages = m.loadHistory()
	} else {
		m.ip = readInput("IP de tu amigo", config.IP)
		m.puerto = readInput("Puerto", config.Port)
		config.Nick = nick; config.IP = m.ip; config.Port = m.puerto; saveConfig(config)
	}

	ti := textinput.New()
	ti.Placeholder = "Escribe un mensaje aquí... (Usa Flechas/PgUp/PgDn para scroll)"
	ti.Focus()
	ti.Prompt = lipgloss.NewStyle().Foreground(purple).Render("> ")
	ti.CharLimit = 120

	// El tamaño interno del viewport calza exacto con el chatBoxStyle
	vp := viewport.New(72, 14)
	m.input = ti
	m.viewport = vp
	m.updateViewport()

	p := text.NewProgram(m, text.WithAltScreen())

	if m.isHost {
		go func() {
			listener, err := net.Listen("tcp", ":"+m.puerto)
			if err != nil {
				p.Send(connectedMsg{err: err})
				return
			}
			conn, err := listener.Accept()
			p.Send(connectedMsg{conn: conn, err: err})
		}()
	} else {
		go func() {
			conn, err := net.Dial("tcp", m.ip+":"+m.puerto)
			p.Send(connectedMsg{conn: conn, err: err})
		}()
	}

	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}

func (m model) Init() text.Cmd {
	return textinput.Blink
}

func (m model) Update(msg text.Msg) (text.Model, text.Cmd) {
	var (
		cmd  text.Cmd
		cmds []text.Cmd
	)

	switch msg := msg.(type) {

	case connectedMsg:
		if msg.err != nil {
			m.messages = append(m.messages, lipgloss.NewStyle().Foreground(red).Render("✘ Error de conexión física."))
			m.updateViewport()
			return m, nil
		}
		m.conn = msg.conn
		m.connected = true

		m.conn.Write([]byte("NICK_HANDSHAKE:" + m.nick + "\n"))

		if m.isHost && len(m.messages) > 0 {
			for _, hMsg := range m.messages {
				if hMsg != "" && !strings.Contains(hMsg, "✔") && !strings.Contains(hMsg, "[!] ") {
					m.conn.Write([]byte("HISTORY_SYNC:" + hMsg + "\n"))
				}
			}
		}

		return m, readNetMessages(m.conn)

	case repoMsg:
		line := string(msg)
		if line == "DISCONNECT_EVENT" {
			m.connected = false
			m.messages = append(m.messages, lipgloss.NewStyle().Foreground(red).Render(fmt.Sprintf("[!] %s ha salido", m.peerNick)))
			m.conn = nil
			m.updateViewport()
			return m, nil
		}

		if strings.HasPrefix(line, "NICK_HANDSHAKE:") {
			m.peerNick = strings.TrimPrefix(line, "NICK_HANDSHAKE:")
			m.messages = append(m.messages, lipgloss.NewStyle().Foreground(green).Render("✔ Conectado con "+m.peerNick))
			m.updateViewport()
			return m, readNetMessages(m.conn)
		}
		if strings.HasPrefix(line, "HISTORY_SYNC:") {
			hist := strings.TrimPrefix(line, "HISTORY_SYNC:")
			m.messages = append(m.messages, hist)
			m.updateViewport()
			return m, readNetMessages(m.conn)
		}

		m.messages = append(m.messages, line)
		m.updateViewport()
		return m, readNetMessages(m.conn)

	case text.KeyMsg:
		switch msg.Type {
		case text.KeyCtrlC:
			if m.conn != nil { m.conn.Close() }
			return m, text.Quit

		// --- CONTROL DE SCROLL IRC-LIKE ---
		case text.KeyUp, text.KeyDown, text.KeyPgUp, text.KeyPgDown:
			// Pasamos las teclas de dirección directo al viewport para que maneje el scroll
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd

		case text.KeyEnter:
			val := strings.TrimSpace(m.input.Value())
			if val == "" { return m, nil }

			if val == "/exit" {
				if m.conn != nil { m.conn.Close() }
				return m, text.Quit
			}
			if val == "/clear" {
				m.messages = []string{}
				m.viewport.SetContent("")
				m.input.SetValue("")
				if m.isHost { os.Remove(m.dbPath) }
				return m, nil
			}

			myLine := myNickStyle.Render(m.nick) + ": " + val
			peerLine := peerStyle.Render(m.nick) + ": " + val

			if m.connected && m.conn != nil {
				m.conn.Write([]byte(peerLine + "\n"))
			}

			m.messages = append(m.messages, myLine)
			m.updateViewport()

			if m.isHost {
				m.saveHistory()
			}

			m.input.SetValue("")
			return m, nil
		}
	}

	// El input sigue capturando el texto normal de las letras
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	return m, text.Batch(cmds...)
}

func (m *model) updateViewport() {
	if len(m.messages) == 0 {
		if m.isHost {
			m.viewport.SetContent(statusStyle.Render("Esperando conexión en puerto " + m.puerto + "... Escribe algo mientras tanto!"))
		} else {
			m.viewport.SetContent(statusStyle.Render("Intentando conectar..."))
		}
	} else {
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
	}
	
	// Auto-scroll al fondo cuando cae un mensaje nuevo
	m.viewport.GotoBottom()
}

func (m model) View() string {
	var header string
	if m.isHost {
		header = titleStyle.Render(" GOChat — Sala Host (DB Activa) ")
	} else {
		header = titleStyle.Render(" GOChat — Cliente ")
	}
	boxContent := chatBoxStyle.Render(m.viewport.View())
	return fmt.Sprintf("%s\n%s\n\n%s", header, boxContent, m.input.View())
}

func (m *model) saveHistory() {
	data, err := json.Marshal(m.messages)
	if err == nil {
		os.WriteFile(m.dbPath, data, 0644)
	}
}

func (m *model) loadHistory() []string {
	var msgs []string
	data, err := os.ReadFile(m.dbPath)
	if err == nil {
		json.Unmarshal(data, &msgs)
		return msgs
	}
	return []string{}
}

func readInput(prompt, def string) string {
	fmt.Printf(" %s [%s]: ", prompt, lipgloss.NewStyle().Foreground(cyan).Render(def))
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" { return def }
	return input
}

func loadConfig() Config {
	var c Config
	c.Nick = "Invitado"; c.IP = "127.0.0.1"; c.Port = "9999"
	data, err := os.ReadFile(configFile)
	if err == nil { json.Unmarshal(data, &c) }
	return c
}

func saveConfig(c Config) {
	data, _ := json.Marshal(c)
	os.WriteFile(configFile, data, 0644)
}
