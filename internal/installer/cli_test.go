package installer_test

import (
	"testing"

	"github.com/your-org/asds-marketplace-setup/internal/installer"
)

func TestCLIInstaller_Method(t *testing.T) {
	inst := installer.NewCLIInstaller("/usr/bin/claude")
	if inst.Method() != "cli" {
		t.Errorf("Method() = %q, want %q", inst.Method(), "cli")
	}
}

func TestCLIInstaller_BuildArgs(t *testing.T) {
	inst := installer.NewCLIInstaller("/usr/bin/claude")

	tests := []struct {
		name string
		fn   func() []string
		want []string
	}{
		{
			"install args",
			func() []string { return inst.BuildInstallArgs("code-reviewer@asds-marketplace", "project") },
			[]string{"plugin", "install", "code-reviewer@asds-marketplace", "--scope", "project"},
		},
		{
			"uninstall args",
			func() []string { return inst.BuildUninstallArgs("code-reviewer@asds-marketplace", "project") },
			[]string{"plugin", "uninstall", "code-reviewer@asds-marketplace", "--scope", "project"},
		},
		{
			"marketplace add args",
			func() []string { return inst.BuildMarketplaceAddArgs("github.com/your-org/asds-marketplace") },
			[]string{"plugin", "marketplace", "add", "github.com/your-org/asds-marketplace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if len(got) != len(tt.want) {
				t.Fatalf("args length = %d, want %d", len(got), len(tt.want))
			}
			for i, a := range got {
				if a != tt.want[i] {
					t.Errorf("args[%d] = %q, want %q", i, a, tt.want[i])
				}
			}
		})
	}
}
