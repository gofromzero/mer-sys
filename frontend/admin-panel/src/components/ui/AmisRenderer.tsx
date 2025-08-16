import { render as amisRender } from 'amis'
import { ToastComponent, AlertComponent } from 'amis-ui'
import axios from 'axios'

interface AmisRendererProps {
  schema: any
  data?: Record<string, any>
}

export const AmisRenderer: React.FC<AmisRendererProps> = ({
  schema,
  data = {}
}) => {
  const amisEnv = {
    // API 请求适配器
    fetcher: async ({ url, method, data: requestData, config }: any) => {
      try {
        const response = await axios({
          url,
          method,
          data: requestData,
          ...config
        })
        return {
          ok: true,
          status: response.status,
          data: response.data
        }
      } catch (error: any) {
        return {
          ok: false,
          status: error.response?.status || 500,
          msg: error.response?.data?.message || error.message
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
      {amisRender(schema, data, amisEnv)}
    </div>
  )
}

export default AmisRenderer