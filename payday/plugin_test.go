package payday

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
	if len(commands) != len(memberCommands) {
		t.Errorf("expected %d commands, got %d", len(memberCommands), len(commands))
	}

	// Check that the payday command is included
	found := false
	for _, cmd := range commands {
		if cmd.Name == "payday" {
			found = true
			break
		}
	}
	if !found {
		t.Error("payday command not found in commands")
	}
}

func TestPluginGetCommandHandlers(t *testing.T) {
	p := &Plugin{}
	handlers := p.GetCommandHandlers()
	if len(handlers) != len(commandHandlers) {
		t.Errorf("expected %d command handlers, got %d", len(commandHandlers), len(handlers))
	}

	// Check that the payday handler is included
	_, found := handlers["payday"]
	if !found {
		t.Error("payday handler not found in command handlers")
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
	if len(help) == 0 {
		t.Error("expected help to not be empty")
	}

	// Check that the help contains the plugin name
	if help[0] != "## Payday\n" {
		t.Errorf("expected first help line to be '## Payday\\n', got '%s'", help[0])
	}
}

func TestPluginGetAdminHelp(t *testing.T) {
	p := &Plugin{}
	adminHelp := p.GetAdminHelp()
	if adminHelp != nil {
		t.Error("expected admin help to be nil")
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
