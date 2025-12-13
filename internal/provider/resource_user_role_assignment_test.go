package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserRoleAssignmentResource(t *testing.T) {
	userID := os.Getenv("ARCHESTRA_TEST_USER_ID")
	if userID == "" {
		t.Skip("ARCHESTRA_TEST_USER_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUserRoleAssignmentResourceConfig(userID, "Test Role Assignment", `["agents:read"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("archestra_user_role_assignment.test", "user_id", userID),
					resource.TestCheckResourceAttrPair("archestra_user_role_assignment.test", "role_id", "archestra_role.test", "id"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccUserRoleAssignmentResourceConfig(userID, roleName, permissions string) string {
	return fmt.Sprintf(`
resource "archestra_role" "test" {
  name        = %[2]q
  permissions = %[3]s
}

resource "archestra_user_role_assignment" "test" {
  user_id = %[1]q
  role_id = archestra_role.test.id
}
`, userID, roleName, permissions)
}
