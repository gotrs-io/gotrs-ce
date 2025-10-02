package playwright

import (
    "testing"
    "time"
    "github.com/gotrs-io/gotrs-ce/tests/e2e/helpers"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestAdminGroupsUI(t *testing.T) {
    browser := helpers.NewBrowserHelper(t)
    if browser.Config.AdminEmail=="" || browser.Config.AdminPassword=="" { t.Skip("Admin credentials not configured") }
    err := browser.Setup(); require.NoError(t, err); defer browser.TearDown()
    auth := helpers.NewAuthHelper(browser)

    t.Run("Admin Groups page loads correctly", func(t *testing.T) {
        err := auth.LoginAsAdmin(); require.NoError(t, err)
        err = browser.NavigateTo("/admin"); require.NoError(t, err); time.Sleep(2*time.Second)
        groupCard := browser.Page.Locator("a[href='/admin/groups']"); c,_ := groupCard.Count(); assert.Greater(t,c,0)
        err = groupCard.Click(); assert.NoError(t, err); time.Sleep(2*time.Second)
        url := browser.Page.URL(); assert.Contains(t,url,"/admin/groups")
        pageTitle := browser.Page.Locator("h1:has-text('Groups')"); c,_ = pageTitle.Count(); assert.Greater(t,c,0)
        addButton := browser.Page.Locator("button:has-text('Add Group')"); c,_ = addButton.Count(); assert.Greater(t,c,0)
        searchInput := browser.Page.Locator("input#groupSearch"); c,_ = searchInput.Count(); assert.Greater(t,c,0)
        groupsTable := browser.Page.Locator("table#groupsTable"); c,_ = groupsTable.Count(); assert.Greater(t,c,0)
        headers := []string{"Group Name","Description","Members","Status","Created"}
        for _,h := range headers { he := browser.Page.Locator("th:has-text('"+h+"')"); c,_ = he.Count(); assert.Greater(t,c,0) }
    })

    t.Run("Add Group modal works", func(t *testing.T) {
        err := browser.NavigateTo("/admin/groups"); require.NoError(t, err); time.Sleep(2*time.Second)
        addButton := browser.Page.Locator("button:has-text('Add Group')"); _ = addButton.Click(); time.Sleep(1*time.Second)
        modal := browser.Page.Locator("#groupModal"); v,_ := modal.IsVisible(); assert.True(t,v)
        nameInput := browser.Page.Locator("input#groupName"); c,_ := nameInput.Count(); assert.Greater(t,c,0)
        descriptionInput := browser.Page.Locator("textarea#groupComments"); c,_ = descriptionInput.Count(); assert.Greater(t,c,0)
        statusSelect := browser.Page.Locator("select#groupStatus"); c,_ = statusSelect.Count(); assert.Greater(t,c,0)
        saveButton := browser.Page.Locator("button:has-text('Save')"); c,_ = saveButton.Count(); assert.Greater(t,c,0)
        cancelButton := browser.Page.Locator("button:has-text('Cancel')"); c,_ = cancelButton.Count(); assert.Greater(t,c,0)
        _ = cancelButton.Click()
    })
}
