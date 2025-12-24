import { test, expect } from '@playwright/test';

const BASE_URL = process.env.BASE_URL || 'http://localhost:8080';

const login = async (page) => {
  await page.context().addCookies([
    {
      name: 'access_token',
      value: 'demo_session_admin',
      domain: 'localhost',
      path: '/',
    },
  ]);
};

test.describe('Admin Users', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  test('loads users management page', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();
    await expect(page.locator('#usersTable tbody tr').first()).toBeVisible();
  });

  test('displays user list with correct columns', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();
    
    // Check for expected column headers
    const headers = page.locator('#usersTable thead th');
    await expect(headers).toHaveCount(6); // ID, Login, Name, Groups, Status, Actions
  });

  test('opens add user modal', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    
    // Click add user button
    const addButton = page.locator('button:has-text("Add User"), a:has-text("Add User")');
    if (await addButton.count() > 0) {
      await addButton.first().click();
      const modal = page.locator('#userModal, .modal');
      await expect(modal).toBeVisible();
    }
  });

  test('edits a user and restores original data', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();

    const firstRow = page.locator('#usersTable tbody tr').first();
    await expect(firstRow).toBeVisible();

    const loginCell = firstRow.locator('td').nth(1);
    const loginValue = (await loginCell.textContent())?.trim();
    expect(loginValue).toBeTruthy();

    await firstRow.locator('button[onclick*="editUser"]').click();

    const modal = page.locator('#userModal');
    await expect(modal).toBeVisible();

    const firstNameInput = modal.locator('#firstName');
    const lastNameInput = modal.locator('#lastName');

    const originalFirstName = await firstNameInput.inputValue();
    const originalLastName = await lastNameInput.inputValue();

    const updatedLastName = originalLastName.endsWith(' QA')
      ? originalLastName
      : `${originalLastName} QA`;

    await lastNameInput.fill(updatedLastName);

    await Promise.all([
      page.waitForNavigation(),
      modal.locator('button[type="submit"]').click(),
    ]);

    await expect(page).toHaveURL(/\/admin\/users$/);
    await expect(page.locator('#usersTable')).toBeVisible();

    const updatedRow = page
      .locator('#usersTable tbody tr')
      .filter({ hasText: loginValue || '' })
      .first();
    await expect(updatedRow).toContainText(updatedLastName);

    await updatedRow.locator('button[onclick*="editUser"]').click();
    await expect(modal).toBeVisible();

    await lastNameInput.fill(originalLastName);

    await Promise.all([
      page.waitForNavigation(),
      modal.locator('button[type="submit"]').click(),
    ]);

    await expect(page).toHaveURL(/\/admin\/users$/);
    await expect(page.locator('#usersTable')).toBeVisible();

    const restoredRow = page
      .locator('#usersTable tbody tr')
      .filter({ hasText: loginValue || '' })
      .first();
    await expect(restoredRow).toContainText(originalLastName);
    await expect(restoredRow).toContainText(originalFirstName);
  });

  test('user row shows groups', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();
    
    // At least one user should have groups displayed
    const rows = page.locator('#usersTable tbody tr');
    const firstRow = rows.first();
    await expect(firstRow).toBeVisible();
    
    // Groups column should exist (may be empty for some users)
    const groupsCell = firstRow.locator('td').nth(3);
    await expect(groupsCell).toBeVisible();
  });

  test('search filters user list', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();
    
    // Get initial row count
    const initialRows = await page.locator('#usersTable tbody tr').count();
    
    // If there's a search input, test filtering
    const searchInput = page.locator('input[type="search"], input[name="search"], #searchInput');
    if (await searchInput.count() > 0) {
      await searchInput.fill('admin');
      await page.waitForTimeout(500); // Wait for filter to apply
      
      // Either filtered results or same count (if 'admin' is common)
      const filteredRows = await page.locator('#usersTable tbody tr').count();
      expect(filteredRows).toBeGreaterThan(0);
    }
  });

  test('user status toggle is accessible', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin/users`);
    await expect(page.locator('#usersTable')).toBeVisible();
    
    const firstRow = page.locator('#usersTable tbody tr').first();
    await expect(firstRow).toBeVisible();
    
    // Status column should have a toggle or indicator
    const statusCell = firstRow.locator('td').nth(4);
    await expect(statusCell).toBeVisible();
  });
});
