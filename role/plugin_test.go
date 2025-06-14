package role

import (
	"testing"

	"github.com/rbrabson/goblin/database/mongo"
	"github.com/rbrabson/goblin/discord"
)

func TestPluginImplementation(t *testing.T) {
	// Ensure the Plugin struct implements the discord.Plugin interface
	var _ discord.Plugin = (*Plugin)(nil)
}

func TestPluginGetName(t *testing.T) {
	p := &Plugin{}
	name := p.GetName()
	if name != PluginName {
		t.Errorf("expected plugin name to be '%s', got '%s'", PluginName, name)
	}
}

func TestPluginGetCommands(t *testing.T) {
	p := &Plugin{}
	commands := p.GetCommands()
	if len(commands) != len(adminCommands) {
		t.Errorf("expected %d commands, got %d", len(adminCommands), len(commands))
	}

	// Check that the guild-admin command is included
	found := false
	for _, cmd := range commands {
		if cmd.Name == "guild-admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("guild-admin command not found in commands")
	}
}

func TestPluginGetCommandHandlers(t *testing.T) {
	p := &Plugin{}
	handlers := p.GetCommandHandlers()
	if len(handlers) != len(commandHandlers) {
		t.Errorf("expected %d command handlers, got %d", len(commandHandlers), len(handlers))
	}

	// Check that the guild-admin handler is included
	_, found := handlers["guild-admin"]
	if !found {
		t.Error("guild-admin handler not found in command handlers")
	}
}

func TestPluginGetComponentHandlers(t *testing.T) {
	p := &Plugin{}
	handlers := p.GetComponentHandlers()
	if handlers != nil {
		t.Error("expected component handlers to be nil")
	}
}

func TestPluginGetHelp(t *testing.T) {
	p := &Plugin{}
	help := p.GetHelp()
	if help != nil {
		t.Error("expected help to be nil")
	}
}

func TestPluginGetAdminHelp(t *testing.T) {
	p := &Plugin{}
	adminHelp := p.GetAdminHelp()
	if len(adminHelp) == 0 {
		t.Error("expected admin help to not be empty")
	}

	// Check that the help contains the plugin name
	if adminHelp[0] != "## Role\n" {
		t.Errorf("expected first help line to be '## Role\\n', got '%s'", adminHelp[0])
	}

	// Check that the help contains the role command
	foundRole := false
	for _, line := range adminHelp {
		if line == "- `/guild-admin role`:  Manages the admin roles for the bot for this server.\n" {
			foundRole = true
			break
		}
	}
	if !foundRole {
		t.Error("expected admin help to contain role command")
	}
}

func TestPluginStatus(t *testing.T) {
	// Save original status
	originalStatus := status
	defer func() {
		status = originalStatus
	}()

	p := &Plugin{}

	// Test RUNNING status
	status = discord.RUNNING
	if p.Status() != discord.RUNNING {
		t.Errorf("expected status to be RUNNING, got %v", p.Status())
	}

	// Test STOPPED status
	status = discord.STOPPED
	if p.Status() != discord.STOPPED {
		t.Errorf("expected status to be STOPPED, got %v", p.Status())
	}
}

func TestPluginStop(t *testing.T) {
	// Save original status
	originalStatus := status
	defer func() {
		status = originalStatus
	}()

	p := &Plugin{}

	// Set status to RUNNING
	status = discord.RUNNING

	// Call Stop
	p.Stop()

	// Check that status is now STOPPED
	if status != discord.STOPPED {
		t.Errorf("expected status to be STOPPED after Stop(), got %v", status)
	}
}

func TestPluginInitialize(t *testing.T) {
	// Save original db
	originalDB := db
	defer func() {
		db = originalDB
	}()

	p := &Plugin{}

	// Set db to nil
	db = nil

	// Create a mock database
	mockDB := &mongo.MongoDB{}

	// Call Initialize
	p.Initialize(nil, mockDB)

	// Check that db is now set to mockDB
	if db != mockDB {
		t.Error("expected db to be set to mockDB")
	}
}

func TestStart(t *testing.T) {
	// This is a simple test to ensure Start doesn't panic
	// We can't easily test that RegisterPlugin is called with the correct plugin
	Start()
	if plugin == nil {
		t.Error("expected plugin to be set")
	}
}