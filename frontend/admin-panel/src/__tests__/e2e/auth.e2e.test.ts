import { test, expect, Page, BrowserContext } from '@playwright/test';

// 测试配置
const BASE_URL = 'http://localhost:5173';
const API_BASE_URL = 'http://localhost:8080/api/v1';

// 测试用户数据
const TEST_USER = {
  username: 'testuser',
  password: 'secret',
  email: 'test@example.com',
  tenant_id: 1
};

const ADMIN_USER = {
  username: 'admin',
  password: 'admin123',
  email: 'admin@example.com',
  tenant_id: 1
};

// 页面对象模式
class LoginPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto(`${BASE_URL}/login`);
  }

  async fillUsername(username: string) {
    await this.page.fill('[data-testid="username-input"]', username);
  }

  async fillPassword(password: string) {
    await this.page.fill('[data-testid="password-input"]', password);
  }

  async checkRememberMe() {
    await this.page.check('[data-testid="remember-me-checkbox"]');
  }

  async clickLogin() {
    await this.page.click('[data-testid="login-button"]');
  }

  async getErrorMessage() {
    return await this.page.textContent('[data-testid="error-message"]');
  }

  async isLoading() {
    return await this.page.isVisible('[data-testid="loading-spinner"]');
  }

  async waitForNavigation() {
    await this.page.waitForURL(`${BASE_URL}/dashboard`);
  }
}

class DashboardPage {
  constructor(private page: Page) {}

  async isVisible() {
    return await this.page.isVisible('[data-testid="dashboard-container"]');
  }

  async getUserInfo() {
    return await this.page.textContent('[data-testid="user-info"]');
  }

  async logout() {
    await this.page.click('[data-testid="logout-button"]');
  }

  async navigateToUsers() {
    await this.page.click('[data-testid="users-menu"]');
  }

  async navigateToAdmin() {
    await this.page.click('[data-testid="admin-menu"]');
  }
}

// 工具函数
async function mockBackendAPI(context: BrowserContext) {
  // 模拟后端API响应
  await context.route(`${API_BASE_URL}/auth/login`, async (route) => {
    const request = route.request();
    const postData = JSON.parse(request.postData() || '{}');
    
    if (postData.username === TEST_USER.username && postData.password === TEST_USER.password) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: '登录成功',
          data: {
            access_token: 'mock-access-token-123',
            refresh_token: 'mock-refresh-token-456',
            user: {
              id: 1,
              username: TEST_USER.username,
              email: TEST_USER.email,
              tenant_id: TEST_USER.tenant_id,
              roles: ['customer'],
              permissions: ['user:view']
            }
          }
        })
      });
    } else {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 40001,
          message: '用户名或密码错误'
        })
      });
    }
  });

  await context.route(`${API_BASE_URL}/auth/logout`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 0,
        message: '登出成功'
      })
    });
  });

  await context.route(`${API_BASE_URL}/auth/refresh`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        code: 0,
        message: '刷新成功',
        data: {
          access_token: 'new-mock-access-token-789',
          refresh_token: 'new-mock-refresh-token-012'
        }
      })
    });
  });

  await context.route(`${API_BASE_URL}/users`, async (route) => {
    const authHeader = route.request().headers()['authorization'];
    if (authHeader && authHeader.includes('mock-access-token')) {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          data: {
            users: [
              { id: 1, username: 'user1', email: 'user1@example.com' },
              { id: 2, username: 'user2', email: 'user2@example.com' }
            ]
          }
        })
      });
    } else {
      await route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 40001,
          message: '未授权访问'
        })
      });
    }
  });
}

test.describe('认证系统端到端测试', () => {
  let loginPage: LoginPage;
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page, context }) => {
    // 设置API模拟
    await mockBackendAPI(context);
    
    loginPage = new LoginPage(page);
    dashboardPage = new DashboardPage(page);
    
    // 清除本地存储
    await page.evaluate(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
  });

  test.describe('登录流程测试', () => {
    test('成功登录流程', async ({ page }) => {
      await loginPage.goto();
      
      // 验证登录页面加载
      await expect(page).toHaveTitle(/登录/);
      await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
      
      // 填写登录信息
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      
      // 点击登录按钮
      await loginPage.clickLogin();
      
      // 验证加载状态
      await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();
      
      // 等待跳转到仪表板
      await loginPage.waitForNavigation();
      
      // 验证登录成功
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
      
      // 验证用户信息显示
      const userInfo = await dashboardPage.getUserInfo();
      expect(userInfo).toContain(TEST_USER.username);
    });

    test('登录失败处理', async ({ page }) => {
      await loginPage.goto();
      
      // 使用错误的凭据
      await loginPage.fillUsername('wronguser');
      await loginPage.fillPassword('wrongpassword');
      await loginPage.clickLogin();
      
      // 验证错误消息显示
      await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
      const errorMessage = await loginPage.getErrorMessage();
      expect(errorMessage).toContain('用户名或密码错误');
      
      // 验证仍在登录页面
      await expect(page).toHaveURL(`${BASE_URL}/login`);
    });

    test('表单验证测试', async ({ page }) => {
      await loginPage.goto();
      
      // 测试空用户名
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      
      await expect(page.locator('[data-testid="username-error"]')).toBeVisible();
      
      // 测试空密码
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword('');
      await loginPage.clickLogin();
      
      await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
      
      // 测试用户名长度验证
      await loginPage.fillUsername('ab'); // 少于3个字符
      await loginPage.fillPassword(TEST_USER.password);
      
      await expect(page.locator('[data-testid="username-error"]')).toBeVisible();
      
      // 测试密码长度验证
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword('12345'); // 少于6个字符
      
      await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
    });

    test('记住我功能测试', async ({ page }) => {
      await loginPage.goto();
      
      // 勾选记住我
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.checkRememberMe();
      await loginPage.clickLogin();
      
      // 等待登录成功
      await loginPage.waitForNavigation();
      
      // 验证localStorage中保存了token
      const savedToken = await page.evaluate(() => localStorage.getItem('access_token'));
      expect(savedToken).toBeTruthy();
      
      // 刷新页面，验证自动登录
      await page.reload();
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
    });
  });

  test.describe('登出流程测试', () => {
    test.beforeEach(async ({ page }) => {
      // 先登录
      await loginPage.goto();
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      await loginPage.waitForNavigation();
    });

    test('成功登出流程', async ({ page }) => {
      // 点击登出
      await dashboardPage.logout();
      
      // 验证跳转到登录页面
      await expect(page).toHaveURL(`${BASE_URL}/login`);
      
      // 验证localStorage被清除
      const token = await page.evaluate(() => localStorage.getItem('access_token'));
      expect(token).toBeNull();
    });

    test('登出后无法访问受保护页面', async ({ page }) => {
      await dashboardPage.logout();
      
      // 尝试直接访问仪表板
      await page.goto(`${BASE_URL}/dashboard`);
      
      // 应该被重定向到登录页面
      await expect(page).toHaveURL(`${BASE_URL}/login`);
    });
  });

  test.describe('权限控制测试', () => {
    test.beforeEach(async ({ page }) => {
      // 登录普通用户
      await loginPage.goto();
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      await loginPage.waitForNavigation();
    });

    test('普通用户访问权限测试', async ({ page }) => {
      // 尝试访问用户列表（有权限）
      await dashboardPage.navigateToUsers();
      await expect(page.locator('[data-testid="users-list"]')).toBeVisible();
      
      // 尝试访问管理页面（无权限）
      await dashboardPage.navigateToAdmin();
      await expect(page.locator('[data-testid="access-denied"]')).toBeVisible();
    });
  });

  test.describe('Token刷新测试', () => {
    test('自动Token刷新', async ({ page, context }) => {
      // 模拟Token即将过期的情况
      await context.route(`${API_BASE_URL}/users`, async (route, request) => {
        const authHeader = request.headers()['authorization'];
        
        if (authHeader && authHeader.includes('mock-access-token-123')) {
          // 第一次请求返回401，触发刷新
          await route.fulfill({
            status: 401,
            contentType: 'application/json',
            body: JSON.stringify({
              code: 40001,
              message: 'Token已过期'
            })
          });
        } else if (authHeader && authHeader.includes('new-mock-access-token-789')) {
          // 刷新后的请求成功
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              code: 0,
              data: { users: [] }
            })
          });
        }
      });
      
      // 先登录
      await loginPage.goto();
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      await loginPage.waitForNavigation();
      
      // 触发需要Token的请求
      await dashboardPage.navigateToUsers();
      
      // 验证页面正常显示（说明Token刷新成功）
      await expect(page.locator('[data-testid="users-list"]')).toBeVisible();
    });
  });

  test.describe('安全性测试', () => {
    test('XSS防护测试', async ({ page }) => {
      await loginPage.goto();
      
      // 尝试注入脚本
      const xssPayload = '<script>alert("XSS")</script>';
      await loginPage.fillUsername(xssPayload);
      await loginPage.fillPassword(TEST_USER.password);
      
      // 验证脚本没有被执行
      const dialogPromise = page.waitForEvent('dialog', { timeout: 1000 }).catch(() => null);
      await loginPage.clickLogin();
      
      const dialog = await dialogPromise;
      expect(dialog).toBeNull(); // 不应该有弹窗
      
      // 验证输入被正确转义
      const usernameValue = await page.inputValue('[data-testid="username-input"]');
      expect(usernameValue).toBe(xssPayload); // 值被保留但不执行
    });

    test('CSRF防护测试', async ({ page, context }) => {
      // 模拟跨站请求
      await context.route(`${API_BASE_URL}/auth/login`, async (route) => {
        const request = route.request();
        const origin = request.headers()['origin'];
        
        if (origin !== BASE_URL) {
          await route.fulfill({
            status: 403,
            contentType: 'application/json',
            body: JSON.stringify({
              code: 40003,
              message: 'CSRF攻击检测'
            })
          });
        } else {
          // 正常请求处理
          await route.continue();
        }
      });
      
      await loginPage.goto();
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      
      // 正常请求应该成功
      await loginPage.waitForNavigation();
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
    });

    test('会话劫持防护测试', async ({ page }) => {
      // 登录获取token
      await loginPage.goto();
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      await loginPage.waitForNavigation();
      
      // 获取当前token
      const originalToken = await page.evaluate(() => localStorage.getItem('access_token'));
      
      // 模拟token被篡改
      await page.evaluate(() => {
        localStorage.setItem('access_token', 'tampered-token-123');
      });
      
      // 刷新页面，应该被重定向到登录页面
      await page.reload();
      await expect(page).toHaveURL(`${BASE_URL}/login`);
    });
  });

  test.describe('响应式设计测试', () => {
    test('移动端登录测试', async ({ page }) => {
      // 设置移动端视口
      await page.setViewportSize({ width: 375, height: 667 });
      
      await loginPage.goto();
      
      // 验证移动端布局
      await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
      
      // 验证表单在移动端正常工作
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      
      await loginPage.waitForNavigation();
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
    });

    test('平板端登录测试', async ({ page }) => {
      // 设置平板端视口
      await page.setViewportSize({ width: 768, height: 1024 });
      
      await loginPage.goto();
      
      // 验证平板端布局和功能
      await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
      
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      await loginPage.clickLogin();
      
      await loginPage.waitForNavigation();
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
    });
  });

  test.describe('性能测试', () => {
    test('登录页面加载性能', async ({ page }) => {
      const startTime = Date.now();
      
      await loginPage.goto();
      
      // 等待页面完全加载
      await page.waitForLoadState('networkidle');
      
      const loadTime = Date.now() - startTime;
      
      // 验证页面在合理时间内加载完成（3秒内）
      expect(loadTime).toBeLessThan(3000);
      
      // 验证关键元素可见
      await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
    });

    test('登录响应时间测试', async ({ page }) => {
      await loginPage.goto();
      
      await loginPage.fillUsername(TEST_USER.username);
      await loginPage.fillPassword(TEST_USER.password);
      
      const startTime = Date.now();
      await loginPage.clickLogin();
      await loginPage.waitForNavigation();
      const responseTime = Date.now() - startTime;
      
      // 验证登录响应时间在合理范围内（5秒内）
      expect(responseTime).toBeLessThan(5000);
    });
  });

  test.describe('可访问性测试', () => {
    test('键盘导航测试', async ({ page }) => {
      await loginPage.goto();
      
      // 使用Tab键导航
      await page.keyboard.press('Tab'); // 聚焦到用户名输入框
      await page.keyboard.type(TEST_USER.username);
      
      await page.keyboard.press('Tab'); // 聚焦到密码输入框
      await page.keyboard.type(TEST_USER.password);
      
      await page.keyboard.press('Tab'); // 聚焦到记住我复选框
      await page.keyboard.press('Space'); // 勾选
      
      await page.keyboard.press('Tab'); // 聚焦到登录按钮
      await page.keyboard.press('Enter'); // 提交表单
      
      await loginPage.waitForNavigation();
      await expect(dashboardPage.isVisible()).resolves.toBe(true);
    });

    test('屏幕阅读器支持测试', async ({ page }) => {
      await loginPage.goto();
      
      // 验证表单标签和aria属性
      await expect(page.locator('[data-testid="username-input"]')).toHaveAttribute('aria-label');
      await expect(page.locator('[data-testid="password-input"]')).toHaveAttribute('aria-label');
      await expect(page.locator('[data-testid="login-button"]')).toHaveAttribute('aria-label');
      
      // 验证错误消息的aria-live属性
      await loginPage.fillUsername('wrong');
      await loginPage.fillPassword('wrong');
      await loginPage.clickLogin();
      
      await expect(page.locator('[data-testid="error-message"]')).toHaveAttribute('aria-live', 'polite');
    });
  });
});