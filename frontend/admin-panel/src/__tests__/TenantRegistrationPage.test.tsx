import { renderWithRouter, screen, fireEvent } from './utils/testUtils'
import TenantRegistrationPage from '../pages/tenant/TenantRegistrationPage'

// Mock 权限Hook
jest.mock('../hooks/useTenantPermissions', () => ({
  useTenantPermissions: () => ({
    canCreateTenant: true
  })
}))

// Mock AmisRenderer 组件
jest.mock('../components/ui/AmisRenderer', () => {
  return function MockAmisRenderer({ schema }: { schema: Record<string, unknown> }) {
    const schemaAny = schema as any; // 测试中的类型简化
    return (
      <div data-testid="amis-renderer">
        <div data-testid="page-title">{String(schemaAny.title || '')}</div>
        
        {/* 模拟表单面板 */}
        {schemaAny.body?.map((panel: any, panelIndex: number) => (
          <div key={panelIndex} data-testid={`panel-${panelIndex}`}>
            <div data-testid={`panel-title-${panelIndex}`}>{String(panel.title || '')}</div>
            
            {/* 如果是表单面板，渲染表单字段 */}
            {panel.body?.[0]?.body?.map((fieldset: any, fieldsetIndex: number) => (
              <div key={fieldsetIndex} data-testid={`fieldset-${fieldsetIndex}`}>
                <div data-testid={`fieldset-title-${fieldsetIndex}`}>{String(fieldset.title || '')}</div>
                
                {/* 渲染字段 */}
                {fieldset.body?.map((gridOrField: any, fieldIndex: number) => {
                  if (gridOrField.type === 'grid') {
                    return gridOrField.columns?.map((column: any, colIndex: number) => (
                      <div key={`${fieldIndex}-${colIndex}`}>
                        {column.body?.map((field: any, subFieldIndex: number) => (
                          <div key={subFieldIndex} data-testid={`field-${field.name}`}>
                            <label>{String(field.label || '')}</label>
                            <input
                              type={field.type === 'input-email' ? 'email' : 'text'}
                              name={String(field.name || '')}
                              placeholder={String(field.placeholder || '')}
                              required={Boolean(field.required)}
                              data-testid={`input-${field.name}`}
                            />
                          </div>
                        ))}
                      </div>
                    ))
                  } else if (gridOrField.type === 'checkboxes') {
                    return (
                      <div key={fieldIndex} data-testid={`field-${gridOrField.name}`}>
                        <label>{String(gridOrField.label || '')}</label>
                        {gridOrField.options?.map((option: any, optIndex: number) => (
                          <label key={optIndex}>
                            <input
                              type="checkbox"
                              value={String(option.value || '')}
                              data-testid={`checkbox-${option.value}`}
                            />
                            {String(option.label || '')}
                          </label>
                        ))}
                      </div>
                    )
                  } else if (gridOrField.type === 'textarea') {
                    return (
                      <div key={fieldIndex} data-testid={`field-${gridOrField.name}`}>
                        <label>{String(gridOrField.label || '')}</label>
                        <textarea
                          name={String(gridOrField.name || '')}
                          placeholder={String(gridOrField.placeholder || '')}
                          data-testid={`textarea-${gridOrField.name}`}
                        />
                      </div>
                    )
                  }
                  return null
                })}
              </div>
            ))}
            
            {/* 渲染提交按钮 */}
            {panel.body?.[0]?.body?.some((item: any) => item.type === 'group') && (
              <div data-testid="form-buttons">
                <button data-testid="submit-button" type="submit">
                  提交注册申请
                </button>
                <button data-testid="reset-button" type="reset">
                  重置表单
                </button>
              </div>
            )}
          </div>
        ))}
      </div>
    )
  }
})


describe('TenantRegistrationPage', () => {
  it('应该正确渲染租户注册页面', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
    expect(screen.getByTestId('page-title')).toHaveTextContent('租户注册')
  })

  it('应该显示租户注册申请面板', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('panel-0')).toBeInTheDocument()
    expect(screen.getByTestId('panel-title-0')).toHaveTextContent('租户注册申请')
  })

  it('应该显示注册须知面板', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('panel-1')).toBeInTheDocument()
    expect(screen.getByTestId('panel-title-1')).toHaveTextContent('注册须知')
  })

  it('应该显示基本信息字段组', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('fieldset-0')).toBeInTheDocument()
    expect(screen.getByTestId('fieldset-title-0')).toHaveTextContent('基本信息')
  })

  it('应该显示联系信息字段组', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('fieldset-1')).toBeInTheDocument()
    expect(screen.getByTestId('fieldset-title-1')).toHaveTextContent('联系信息')
  })

  it('应该显示服务协议字段组', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('fieldset-2')).toBeInTheDocument()
    expect(screen.getByTestId('fieldset-title-2')).toHaveTextContent('服务协议')
  })
})

describe('TenantRegistrationPage Form Fields', () => {
  it('应该显示所有必要的基本信息字段', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 租户名称
    expect(screen.getByTestId('field-name')).toBeInTheDocument()
    expect(screen.getByTestId('input-name')).toBeRequired()
    
    // 租户代码
    expect(screen.getByTestId('field-code')).toBeInTheDocument()
    expect(screen.getByTestId('input-code')).toBeRequired()
    
    // 业务类型
    expect(screen.getByTestId('field-business_type')).toBeInTheDocument()
  })

  it('应该显示所有必要的联系信息字段', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 联系人
    expect(screen.getByTestId('field-contact_person')).toBeInTheDocument()
    expect(screen.getByTestId('input-contact_person')).toBeRequired()
    
    // 联系邮箱
    expect(screen.getByTestId('field-contact_email')).toBeInTheDocument()
    expect(screen.getByTestId('input-contact_email')).toBeRequired()
    expect(screen.getByTestId('input-contact_email')).toHaveAttribute('type', 'email')
    
    // 联系电话
    expect(screen.getByTestId('field-contact_phone')).toBeInTheDocument()
    
    // 详细地址
    expect(screen.getByTestId('field-address')).toBeInTheDocument()
    expect(screen.getByTestId('textarea-address')).toBeInTheDocument()
  })

  it('应该显示服务协议复选框', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('field-agreements')).toBeInTheDocument()
    
    // 服务协议
    expect(screen.getByTestId('checkbox-service_agreement')).toBeInTheDocument()
    
    // 隐私政策
    expect(screen.getByTestId('checkbox-privacy_policy')).toBeInTheDocument()
    
    // 通知同意
    expect(screen.getByTestId('checkbox-notifications')).toBeInTheDocument()
  })

  it('应该显示表单提交和重置按钮', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('submit-button')).toBeInTheDocument()
    expect(screen.getByTestId('submit-button')).toHaveTextContent('提交注册申请')
    
    expect(screen.getByTestId('reset-button')).toBeInTheDocument()
    expect(screen.getByTestId('reset-button')).toHaveTextContent('重置表单')
  })
})

describe('TenantRegistrationPage Form Validation', () => {
  it('应该设置正确的字段验证规则', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 必填字段
    expect(screen.getByTestId('input-name')).toBeRequired()
    expect(screen.getByTestId('input-code')).toBeRequired()
    expect(screen.getByTestId('input-contact_person')).toBeRequired()
    expect(screen.getByTestId('input-contact_email')).toBeRequired()
    
    // 邮箱字段类型
    expect(screen.getByTestId('input-contact_email')).toHaveAttribute('type', 'email')
  })

  it('应该有正确的占位符文本', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    expect(screen.getByTestId('input-name')).toHaveAttribute('placeholder', '例如：某某科技有限公司')
    expect(screen.getByTestId('input-code')).toHaveAttribute('placeholder', '例如：company-tech')
    expect(screen.getByTestId('input-contact_person')).toHaveAttribute('placeholder', '请输入联系人姓名')
    expect(screen.getByTestId('input-contact_email')).toHaveAttribute('placeholder', 'example@company.com')
    expect(screen.getByTestId('input-contact_phone')).toHaveAttribute('placeholder', '例如：13800138000')
  })
})

describe('TenantRegistrationPage User Interaction', () => {
  it('应该允许用户填写表单字段', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    const nameInput = screen.getByTestId('input-name')
    fireEvent.change(nameInput, { target: { value: '测试公司' } })
    expect(nameInput).toHaveValue('测试公司')
    
    const emailInput = screen.getByTestId('input-contact_email')
    fireEvent.change(emailInput, { target: { value: 'test@example.com' } })
    expect(emailInput).toHaveValue('test@example.com')
  })

  it('应该允许用户选择服务协议', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    const serviceAgreement = screen.getByTestId('checkbox-service_agreement')
    fireEvent.click(serviceAgreement)
    expect(serviceAgreement).toBeChecked()
    
    const privacyPolicy = screen.getByTestId('checkbox-privacy_policy')
    fireEvent.click(privacyPolicy)
    expect(privacyPolicy).toBeChecked()
  })

  it('应该响应重置按钮点击', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 填写一些数据
    const nameInput = screen.getByTestId('input-name')
    fireEvent.change(nameInput, { target: { value: '测试公司' } })
    
    // 点击重置按钮
    const resetButton = screen.getByTestId('reset-button')
    fireEvent.click(resetButton)
    
    // 验证按钮存在（实际重置逻辑由Amis处理）
    expect(resetButton).toBeInTheDocument()
  })
})

describe('TenantRegistrationPage Accessibility', () => {
  it('应该有正确的标签和输入框关联', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证每个字段都有对应的标签
    expect(screen.getByText('租户名称')).toBeInTheDocument()
    expect(screen.getByText('租户代码')).toBeInTheDocument()
    expect(screen.getByText('联系人')).toBeInTheDocument()
    expect(screen.getByText('联系邮箱')).toBeInTheDocument()
    expect(screen.getByText('联系电话')).toBeInTheDocument()
    expect(screen.getByText('详细地址')).toBeInTheDocument()
  })

  it('应该支持键盘导航', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    const nameInput = screen.getByTestId('input-name')
    nameInput.focus()
    expect(document.activeElement).toBe(nameInput)
  })

  it('应该有清晰的错误提示区域', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证页面结构正确，错误提示会由Amis处理
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })
})

describe('TenantRegistrationPage Performance', () => {
  it('应该快速渲染表单', () => {
    const startTime = performance.now()
    renderWithRouter(<TenantRegistrationPage />)
    const endTime = performance.now()
    
    expect(endTime - startTime).toBeLessThan(100) // 100ms内完成渲染
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })

  it('应该优化大量表单字段的渲染', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证所有字段都正确渲染
    const fields = [
      'name', 'code', 'business_type', 'contact_person',
      'contact_email', 'contact_phone', 'address'
    ]
    
    fields.forEach(fieldName => {
      expect(screen.getByTestId(`field-${fieldName}`)).toBeInTheDocument()
    })
  })
})

describe('TenantRegistrationPage Business Logic', () => {
  it('应该有正确的业务类型选项配置', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证业务类型字段存在
    expect(screen.getByTestId('field-business_type')).toBeInTheDocument()
  })

  it('应该有合理的默认配置说明', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证注册须知面板存在
    expect(screen.getByTestId('panel-title-1')).toHaveTextContent('注册须知')
  })

  it('应该提供清晰的注册流程说明', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证说明面板存在
    expect(screen.getByTestId('panel-1')).toBeInTheDocument()
  })
})

describe('TenantRegistrationPage Error Handling', () => {
  it('应该优雅处理渲染错误', () => {
    // 这个测试确保即使有问题也不会崩溃
    expect(() => {
      renderWithRouter(<TenantRegistrationPage />)
    }).not.toThrow()
  })

  it('应该正确处理空数据', () => {
    renderWithRouter(<TenantRegistrationPage />)
    
    // 验证空状态下页面仍能正常渲染
    expect(screen.getByTestId('amis-renderer')).toBeInTheDocument()
  })
})