#!/usr/bin/env node

const https = require('https');
const http = require('http');

// 测试配置
const BASE_URL = 'http://localhost:8081';
const TEST_CREDENTIALS = {
  username: 'admin',
  password: 'password123',
  tenant_id: 1
};

// HTTP请求辅助函数
function makeRequest(options, data = null) {
  return new Promise((resolve, reject) => {
    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => {
        body += chunk;
      });
      res.on('end', () => {
        try {
          const parsedBody = body ? JSON.parse(body) : {};
          resolve({
            statusCode: res.statusCode,
            headers: res.headers,
            data: parsedBody
          });
        } catch (error) {
          resolve({
            statusCode: res.statusCode,
            headers: res.headers,
            data: body
          });
        }
      });
    });

    req.on('error', (error) => {
      reject(error);
    });

    if (data) {
      req.write(JSON.stringify(data));
    }
    req.end();
  });
}

// 测试用例
async function testAuthFlow() {
  console.log('🚀 开始端到端认证流程测试...\n');

  try {
    // 1. 测试健康检查
    console.log('1. 测试健康检查端点...');
    const healthResponse = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/health',
      method: 'GET'
    });
    
    if (healthResponse.statusCode === 200) {
      console.log('✅ 健康检查成功:', healthResponse.data);
    } else {
      console.log('❌ 健康检查失败:', healthResponse.statusCode);
      return;
    }

    // 2. 测试登录端点（先测试错误情况）
    console.log('\n2. 测试无效登录...');
    const invalidLoginResponse = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/api/v1/auth/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, {
      username: 'invalid',
      password: 'invalid'
    });

    console.log('无效登录响应:', invalidLoginResponse.statusCode, invalidLoginResponse.data);

    // 3. 测试有效登录（如果用户存在）
    console.log('\n3. 测试登录端点...');
    const loginResponse = await makeRequest({
      hostname: 'localhost',
      port: 8081,
      path: '/api/v1/auth/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }, TEST_CREDENTIALS);

    console.log('登录响应状态:', loginResponse.statusCode);
    console.log('登录响应数据:', loginResponse.data);

    let accessToken = null;
    let refreshToken = null;

    if (loginResponse.statusCode === 200 && loginResponse.data.data) {
      accessToken = loginResponse.data.data.access_token;
      refreshToken = loginResponse.data.data.refresh_token;
      console.log('✅ 登录成功，获得令牌');
    } else {
      console.log('⚠️  登录响应非标准格式，可能是因为用户不存在或其他原因');
    }

    // 4. 测试受保护的端点（如果有令牌）
    if (accessToken) {
      console.log('\n4. 测试受保护的用户信息端点...');
      const userInfoResponse = await makeRequest({
        hostname: 'localhost',
        port: 8081,
        path: '/api/v1/user/info',
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        }
      });

      console.log('用户信息响应:', userInfoResponse.statusCode, userInfoResponse.data);

      // 5. 测试令牌刷新
      if (refreshToken) {
        console.log('\n5. 测试令牌刷新...');
        const refreshResponse = await makeRequest({
          hostname: 'localhost',
          port: 8081,
          path: '/api/v1/auth/refresh',
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          }
        }, {
          refresh_token: refreshToken
        });

        console.log('令牌刷新响应:', refreshResponse.statusCode, refreshResponse.data);
      }

      // 6. 测试登出
      console.log('\n6. 测试登出...');
      const logoutResponse = await makeRequest({
        hostname: 'localhost',
        port: 8081,
        path: '/api/v1/auth/logout',
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${accessToken}`,
          'Content-Type': 'application/json'
        }
      });

      console.log('登出响应:', logoutResponse.statusCode, logoutResponse.data);
    }

    console.log('\n🎉 端到端测试完成！');

  } catch (error) {
    console.error('❌ 测试失败:', error.message);
  }
}

// 前端连接测试
async function testFrontendIntegration() {
  console.log('\n🌐 测试前端集成...');
  
  try {
    // 检查前端是否在运行
    const frontendRequest = http.request({
      hostname: 'localhost',
      port: 5173, // Vite默认端口
      path: '/',
      method: 'HEAD'
    }, (res) => {
      if (res.statusCode === 200) {
        console.log('✅ 前端服务运行正常 (端口5173)');
      } else {
        console.log('⚠️  前端服务响应异常:', res.statusCode);
      }
    });

    frontendRequest.on('error', () => {
      console.log('❌ 前端服务未运行 (端口5173)');
      console.log('💡 提示: 运行 `cd frontend/admin-panel && npm run dev` 启动前端服务');
    });

    frontendRequest.setTimeout(2000, () => {
      console.log('⏰ 前端服务连接超时');
    });

    frontendRequest.end();

  } catch (error) {
    console.log('❌ 前端集成测试失败:', error.message);
  }
}

// 数据库连接测试
async function testDatabaseConnection() {
  console.log('\n💾 测试数据库连接...');
  
  // 这里我们通过API间接测试数据库连接
  // 实际的数据库连接测试应该在后端代码中进行
  console.log('💡 数据库连接测试需要通过后端API间接进行');
  console.log('💡 如果登录/用户信息等API正常工作，说明数据库连接正常');
}

// 执行所有测试
async function runAllTests() {
  console.log('🧪 MER系统认证流程端到端测试');
  console.log('=====================================\n');

  await testAuthFlow();
  await testFrontendIntegration();
  await testDatabaseConnection();

  console.log('\n📋 测试总结:');
  console.log('- 后端用户服务: http://localhost:8081');
  console.log('- 前端管理面板: http://localhost:5173 (如果运行)');
  console.log('- MySQL数据库: localhost:3306');
  console.log('- Redis缓存: localhost:6379');
  console.log('\n✨ 端到端测试执行完毕！');
}

// 运行测试
runAllTests().catch(console.error);