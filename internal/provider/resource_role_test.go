package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRoleResourceConfig("Test Role", "Test Description", `["agents:read", "agents:write"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("archestra_role.test", "name", "Test Role"),
					resource.TestCheckResourceAttr("archestra_role.test", "description", "Test Description"),
					resource.TestCheckResourceAttr("archestra_role.test", "permissions.#", "2"),
					resource.TestCheckResourceAttr("archestra_role.test", "permissions.0", "agents:read"),
					resource.TestCheckResourceAttr("archestra_role.test", "permissions.1", "agents:write"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "archestra_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccRoleResourceConfig("Updated Role", "Updated Description", `["mcp_servers:read"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("archestra_role.test", "name", "Updated Role"),
					resource.TestCheckResourceAttr("archestra_role.test", "description", "Updated Description"),
					resource.TestCheckResourceAttr("archestra_role.test", "permissions.#", "1"),
					resource.TestCheckResourceAttr("archestra_role.test", "permissions.0", "mcp_servers:read"),
				),
			},
			// Data Source testing
			{
				Config: testAccRoleResourceConfig("Updated Role", "Updated Description", `["mcp_servers:read"]`) + `
					data "archestra_role" "test" {
						id = archestra_role.test.id
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.archestra_role.test", "name", "archestra_role.test", "name"),
					resource.TestCheckResourceAttrPair("data.archestra_role.test", "description", "archestra_role.test", "description"),
					resource.TestCheckResourceAttrPair("data.archestra_role.test", "permissions.#", "archestra_role.test", "permissions.#"),
				),
			},
		},
	})
}

func testAccRoleResourceConfig(name, description, permissions string) string {
	return fmt.Sprintf(`
resource "archestra_role" "test" {
  name        = %[1]q
  description = %[2]q
  permissions = %[3]s
}
`, name, description, permissions)
}
