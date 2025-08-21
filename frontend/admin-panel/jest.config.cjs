/** @type {import('jest').Config} */
const config = {
  // 测试环境
  testEnvironment: '@happy-dom/jest-environment',
  
  // 根目录
  rootDir: '.',
  
  // 测试文件匹配模式
  testMatch: [
    '<rootDir>/src/**/__tests__/**/*.{js,jsx,ts,tsx}',
    '<rootDir>/src/**/*.(test|spec).{js,jsx,ts,tsx}'
  ],
  
  // 模块文件扩展名
  moduleFileExtensions: ['js', 'jsx', 'ts', 'tsx', 'json'],
  
  // 转换配置
  transform: {
    '^.+\\.(ts|tsx)$': ['ts-jest', {
      tsconfig: {
        jsx: 'react-jsx',
        esModuleInterop: true,
        allowSyntheticDefaultImports: true
      }
    }],
    '^.+\\.(js|jsx)$': 'babel-jest'
  },
  
  // 模块名映射（处理路径别名）
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/src/$1',
    '\\.(css|less|scss|sass)$': 'identity-obj-proxy',
    '^canvas$': '<rootDir>/src/__mocks__/canvas.js'
  },
  
  // 设置文件
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  
  // 覆盖率配置
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/main.tsx',
    '!src/vite-env.d.ts'
  ],
  
  // 覆盖率阈值
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  },
  
  // 覆盖率报告格式
  coverageReporters: ['text', 'lcov', 'html'],
  
  // 忽略的路径
  testPathIgnorePatterns: [
    '<rootDir>/node_modules/',
    '<rootDir>/dist/'
  ],
  
  // 模块路径忽略
  modulePathIgnorePatterns: ['<rootDir>/dist/'],
  
  // 清除模拟
  clearMocks: true,
  
  // 恢复模拟
  restoreMocks: true,
  
  // 测试超时
  testTimeout: 10000,
  
  // 全局变量
  globals: {}
};

module.exports = config;