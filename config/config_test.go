package config

import "testing"

func TestConfigToolArchiveType_IsValid(t *testing.T) {
	testTable := []struct {
		name  string
		valid bool
	}{
		{name: "aaa", valid: false},
		{name: "tar.gz", valid: true},
	}
	for _, tv := range testTable {
		t.Run(tv.name, func(t *testing.T) {
			archiveType := ConfigToolArchiveType(tv.name)
			if archiveType.IsValid() != tv.valid {
				t.Fatalf("expecting check to pass")
			}
		})
	}
}

func TestConfigToolArchiveType_IsTar(t *testing.T) {
	testTable := []struct {
		name  ConfigToolArchiveType
		valid bool
	}{
		{name: ConfigToolArchiveType("aaa"), valid: false},
		{name: ArchiveTypeZip, valid: false},
		{name: ArchiveTypeTarGz, valid: true},
	}
	for _, tv := range testTable {
		t.Run(string(tv.name), func(t *testing.T) {
			if tv.name.IsTar() != tv.valid {
				t.Fatalf("expecting check to pass")
			}
		})
	}
}
