import React from 'react'
import { render } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'

// 带有 React Router 的测试渲染工具
export const renderWithRouter = (component: React.ReactElement) => {
  return render(
    React.createElement(BrowserRouter, { future: { v7_startTransition: true } }, component)
  )
}

// 导出供测试使用的通用工具
export * from '@testing-library/react'
export { screen, fireEvent, waitFor } from '@testing-library/react'

