package store

import (
	"testing"
)

func TestServerCRUD(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	srv := &Server{
		ID: "test-1", Name: "Test Server", Host: "1.2.3.4",
		Port: 22, Username: "root", AuthMethod: "password",
		Credential: "secret", Status: "unknown", Source: "fresh",
		Tags: "[]",
	}

	if err := s.CreateServer(srv); err != nil {
		t.Fatalf("CreateServer: %v", err)
	}

	got, err := s.GetServer("test-1")
	if err != nil {
		t.Fatalf("GetServer: %v", err)
	}
	if got.Name != "Test Server" {
		t.Errorf("name = %q, want %q", got.Name, "Test Server")
	}

	servers, err := s.ListServers()
	if err != nil {
		t.Fatalf("ListServers: %v", err)
	}
	if len(servers) != 1 {
		t.Errorf("len = %d, want 1", len(servers))
	}

	if err := s.DeleteServer("test-1"); err != nil {
		t.Fatalf("DeleteServer: %v", err)
	}

	_, err = s.GetServer("test-1")
	if err == nil {
		t.Error("expected error for deleted server")
	}
}
