import { render as amisRender, type Schema } from 'amis'
import { ToastComponent, AlertComponent } from 'amis-ui'
import axios from 'axios'

interface AmisRendererProps {
  schema: Schema
  data?: Record<string, unknown>
}

export const AmisRenderer: React.FC<AmisRendererProps> = ({
  schema,
  data = {}
}) => {
  const amisEnv = {
    // API 请求适配器
    fetcher: async (config: Record<string, unknown>) => {
      const { url, method = 'GET', data: requestData, ...restConfig } = config
      try {
        const response = await axios({
          url: url as string,
          method: method as string,
          data: requestData,
          ...(restConfig as Record<string, unknown>)
        })
        return {
          ok: true,
          status: response.status,
          data: response.data
        }
      } catch (error: unknown) {
        const axiosError = error as { response?: { status?: number; data?: { message?: string } }; message?: string }
        return {
          ok: false,
          status: axiosError.response?.status || 500,
          msg: axiosError.response?.data?.message || axiosError.message
        }
      }
    },

    // 是否为移动端
    isMobile: () => false,

    // 跳转适配器
    jumpTo: (to: string) => {
      // 这里可以集成 React Router 的导航
      window.location.hash = to
    },

    // 更新地址栏
    updateLocation: (to: string) => {
      window.history.pushState({}, '', to)
    },

    // 是否为 IE
    isCancel: () => false,

    // Toast 提示
    notify: (type: string, msg: string) => {
      ToastComponent.show({
        message: msg,
        level: type
      })
    },

    // Alert 确认
    alert: (content: string) => {
      AlertComponent.show({
        content,
        confirmText: '确定'
      })
    },

    // 确认框
    confirm: (content: string): Promise<boolean> => {
      return new Promise<boolean>((resolve) => {
        AlertComponent.show({
          content,
          confirmText: '确定',
          cancelText: '取消',
          onConfirm: () => resolve(true),
          onCancel: () => resolve(false)
        })
      })
    }
  }

  return (
    <div className="amis-scope">
      {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
      {amisRender(schema, data, amisEnv as any)}
    </div>
  )
}

export default AmisRenderer