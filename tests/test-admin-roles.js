const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch({ 
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  });
  
  const context = await browser.newContext({
    ignoreHTTPSErrors: true
  });
  
  const page = await context.newPage();
  
  try {
    console.log('🔍 Testing Admin Roles Page...\n');
    
    // Navigate to login
    console.log('1. Navigating to login page...');
    await page.goto('http://nginx/login');
    
    // Login
    console.log('2. Logging in as admin...');
    await page.fill('input[name="username"]', 'root@localhost');
    await page.fill('input[name="password"]', 'root');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard', { timeout: 5000 });
    console.log('   ✅ Login successful');
    
    // Navigate to roles
    console.log('\n3. Navigating to Admin Roles...');
    await page.goto('http://nginx/admin/roles');
    await page.waitForSelector('table', { timeout: 5000 });
    console.log('   ✅ Roles page loaded');
    
    // Check roles list
    console.log('\n4. Checking roles list...');
    const roles = await page.$$eval('tbody tr', rows => 
      rows.map(row => ({
        name: row.querySelector('td:first-child')?.textContent?.trim(),
        status: row.querySelector('td:nth-child(4) span')?.textContent?.trim()
      }))
    );
    
    console.log('   Found roles:');
    roles.forEach(role => {
      console.log(`   - ${role.name}: ${role.status}`);
    });
    
    // Test edit functionality
    console.log('\n5. Testing edit functionality...');
    await page.click('tbody tr:first-child button[title="Edit role"]');
    await page.waitForSelector('#roleModal:not(.hidden)', { timeout: 5000 });
    
    const roleName = await page.inputValue('#roleName');
    console.log(`   ✅ Edit modal opened for: ${roleName}`);
    
    // Check permissions
    const checkedPermissions = await page.$$eval(
      'input[name="permissions"]:checked', 
      checkboxes => checkboxes.map(cb => cb.value)
    );
    console.log(`   Permissions: ${checkedPermissions.length > 0 ? checkedPermissions.join(', ') : 'None'}`);
    
    // Close modal
    await page.click('#roleModal button[onclick*="closeRoleModal"]');
    
    // Test membership management
    console.log('\n6. Testing membership management...');
    await page.click('tbody tr:first-child button[title="Manage users"]');
    await page.waitForSelector('#roleUsersModal:not(.hidden)', { timeout: 5000 });
    console.log('   ✅ Membership modal opened');
    
    // Check available users
    const availableUsers = await page.$$eval(
      '#availableUsersList li', 
      items => items.length
    );
    console.log(`   Available users: ${availableUsers}`);
    
    // Check current members
    const currentMembers = await page.$$eval(
      '#currentMembersList li', 
      items => items.length
    );
    console.log(`   Current members: ${currentMembers}`);
    
    await page.screenshot({ 
      path: '/test-results/admin-roles-membership.png',
      fullPage: true 
    });
    console.log('   📸 Screenshot saved: admin-roles-membership.png');
    
    console.log('\n✅ All tests passed!');
    
  } catch (error) {
    console.error('\n❌ Test failed:', error.message);
    await page.screenshot({ 
      path: '/test-results/error-screenshot.png',
      fullPage: true 
    });
    process.exit(1);
  } finally {
    await browser.close();
  }
})();