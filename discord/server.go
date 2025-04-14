package discord

import (
	"slices"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Server represents the owners and admins of the bot in the database.
type Server struct {
	ID     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Owners []string           `json:"owners" bson:"owners"`
	Admins []string           `json:"admins" bson:"admins"`
}

// GetServer retrieves the bot from the database.
func GetServer() *Server {
	server := ReadServer()
	if server == nil {
		server = NewServer()
	}
	return server
}

// NewServer creates a new server in the database, and writes the newly created server to the database.
func NewServer() *Server {
	server := &Server{
		Owners: []string{},
		Admins: []string{},
	}
	WriteServer(server)
	return server
}

// AddOwner adds a member as an owner of the server.
func (s *Server) AddOwner(memberID string) error {
	if slices.Contains(s.Owners, memberID) {
		return ErrAlreadyOwner
	}
	s.Owners = append(s.Owners, memberID)
	return WriteServer(s)
}

// RemoveOwner removes a member as an owner of the server.
func (s *Server) RemoveOwner(memberID string) error {
	if !slices.Contains(s.Owners, memberID) {
		return ErrNotOwner
	}
	s.Owners = slices.DeleteFunc(s.Owners, func(s string) bool {
		return s == memberID
	})
	return WriteServer(s)
}

// ListOwners lists the owners of the server.
func (s *Server) ListOwners() []string {
	return s.Owners
}

// AddAdmin adds a member as an admin of the server.
func (s *Server) AddAdmin(memberID string) error {
	if slices.Contains(s.Admins, memberID) {
		return ErrAlreadyAdmin
	}
	s.Admins = append(s.Admins, memberID)
	return WriteServer(s)
}

// RemoveAdmin removes a member as an admin of the server.
func (s *Server) RemoveAdmin(memberID string) error {
	if !slices.Contains(s.Admins, memberID) {
		return ErrNotAdmin
	}
	s.Owners = slices.DeleteFunc(s.Owners, func(s string) bool {
		return s == memberID
	})
	return WriteServer(s)
}

// ListAdmins lists the admins of the server.
func (s *Server) ListAdmins() []string {
	return s.Admins
}

// HasOwners checks if the server has any owners.
func (s *Server) HasOwners() bool {
	return len(s.Owners) > 0
}

// IsOwner checks if the given member is an owner of the server.
func (s *Server) IsOwner(memberID string) bool {
	return slices.Contains(s.Owners, memberID)
}

// HasAdmins checks if the server has any admins.
func (s *Server) HasAdmins() bool {
	return len(s.Admins) > 0
}

// IsAdmin checks if the given member is an admin of the server.
func (s *Server) IsAdmin(memberID string) bool {
	return slices.Contains(s.Admins, memberID)
}
