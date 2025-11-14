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
});
