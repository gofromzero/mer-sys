import React, { useState, useEffect } from 'react';
import { Plus, Play, Pause, Edit, Trash2, Clock, Calendar, AlertCircle, CheckCircle, XCircle } from 'lucide-react';

interface ScheduledTask {
  id: number;
  task_name: string;
  task_description: string;
  report_type: string;
  cron_expression: string;
  report_config: string;
  recipients: string[];
  is_enabled: boolean;
  last_run_time?: string;
  last_run_status: string;
  last_run_message: string;
  next_run_time?: string;
  created_at: string;
  updated_at: string;
}

interface CreateTaskRequest {
  task_name: string;
  task_description: string;
  report_type: string;
  cron_expression: string;
  report_config: string;
  recipients: string[];
  is_enabled: boolean;
}

interface UpdateTaskRequest {
  task_name?: string;
  task_description?: string;
  cron_expression?: string;
  report_config?: string;
  recipients?: string[];
  is_enabled?: boolean;
}

const ScheduledTaskPage: React.FC = () => {
  const [tasks, setTasks] = useState<ScheduledTask[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingTask, setEditingTask] = useState<ScheduledTask | null>(null);
  const [executing, setExecuting] = useState<Set<number>>(new Set());

  // 分页状态
  const [page, setPage] = useState(1);
  const [pageSize] = useState(10);
  const [total, setTotal] = useState(0);

  // 筛选状态
  const [filters, setFilters] = useState({
    task_name: '',
    report_type: '',
    is_enabled: undefined as boolean | undefined,
    last_run_status: '',
  });

  // 加载定时任务列表
  const loadTasks = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const params = new URLSearchParams({
        page: page.toString(),
        page_size: pageSize.toString(),
        ...Object.fromEntries(
          Object.entries(filters).filter(([_, value]) => value !== '' && value !== undefined)
        ),
      });

      const response = await fetch(`/api/v1/scheduled-tasks?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
      });

      if (!response.ok) {
        throw new Error('获取定时任务列表失败');
      }

      const data = await response.json();
      setTasks(data.data.tasks || []);
      setTotal(data.data.pagination.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : '未知错误');
    } finally {
      setLoading(false);
    }
  };

  // 创建定时任务
  const createTask = async (taskData: CreateTaskRequest) => {
    try {
      const response = await fetch('/api/v1/scheduled-tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
        body: JSON.stringify(taskData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || '创建定时任务失败');
      }

      await loadTasks();
      setShowCreateModal(false);
    } catch (err) {
      throw err;
    }
  };

  // 更新定时任务
  const updateTask = async (taskId: number, taskData: UpdateTaskRequest) => {
    try {
      const response = await fetch(`/api/v1/scheduled-tasks/${taskId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
        body: JSON.stringify(taskData),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || '更新定时任务失败');
      }

      await loadTasks();
      setShowEditModal(false);
      setEditingTask(null);
    } catch (err) {
      throw err;
    }
  };

  // 删除定时任务
  const deleteTask = async (taskId: number) => {
    if (!confirm('确定要删除这个定时任务吗？')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/scheduled-tasks/${taskId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
      });

      if (!response.ok) {
        throw new Error('删除定时任务失败');
      }

      await loadTasks();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除失败');
    }
  };

  // 切换任务启用状态
  const toggleTask = async (taskId: number, enabled: boolean) => {
    try {
      const response = await fetch(`/api/v1/scheduled-tasks/${taskId}/toggle`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
        body: JSON.stringify({ enabled }),
      });

      if (!response.ok) {
        throw new Error('切换任务状态失败');
      }

      await loadTasks();
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败');
    }
  };

  // 手动执行任务
  const executeTask = async (taskId: number) => {
    setExecuting(prev => new Set(prev).add(taskId));
    
    try {
      const response = await fetch(`/api/v1/scheduled-tasks/${taskId}/execute`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'X-Tenant-ID': localStorage.getItem('tenantId') || '1',
        },
      });

      if (!response.ok) {
        throw new Error('执行任务失败');
      }

      await loadTasks();
    } catch (err) {
      setError(err instanceof Error ? err.message : '执行失败');
    } finally {
      setExecuting(prev => {
        const newSet = new Set(prev);
        newSet.delete(taskId);
        return newSet;
      });
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-500" />;
      case 'running':
        return <Clock className="w-4 h-4 text-blue-500 animate-spin" />;
      case 'pending':
      default:
        return <AlertCircle className="w-4 h-4 text-gray-400" />;
    }
  };

  // 获取状态文字
  const getStatusText = (status: string) => {
    const statusMap: Record<string, string> = {
      pending: '待执行',
      running: '运行中',
      completed: '已完成',
      failed: '失败',
    };
    return statusMap[status] || status;
  };

  // 格式化日期时间
  const formatDateTime = (dateTime?: string) => {
    if (!dateTime) return '-';
    return new Date(dateTime).toLocaleString('zh-CN');
  };

  // 获取报表类型中文名
  const getReportTypeName = (type: string) => {
    const typeMap: Record<string, string> = {
      financial: '财务报表',
      merchant_operation: '商户运营',
      customer_behavior: '客户行为',
    };
    return typeMap[type] || type;
  };

  useEffect(() => {
    loadTasks();
  }, [page, filters]);

  return (
    <div className="scheduled-task-page p-6">
      {/* 页面标题和操作按钮 */}
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">定时任务管理</h1>
          <p className="mt-1 text-sm text-gray-600">管理和监控自动化报表任务</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
        >
          <Plus className="w-4 h-4 mr-2" />
          创建任务
        </button>
      </div>

      {/* 筛选器 */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">任务名称</label>
            <input
              type="text"
              value={filters.task_name}
              onChange={(e) => setFilters(prev => ({ ...prev, task_name: e.target.value }))}
              placeholder="搜索任务名称"
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">报表类型</label>
            <select
              value={filters.report_type}
              onChange={(e) => setFilters(prev => ({ ...prev, report_type: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">全部类型</option>
              <option value="financial">财务报表</option>
              <option value="merchant_operation">商户运营</option>
              <option value="customer_behavior">客户行为</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">启用状态</label>
            <select
              value={filters.is_enabled === undefined ? '' : filters.is_enabled.toString()}
              onChange={(e) => setFilters(prev => ({ 
                ...prev, 
                is_enabled: e.target.value === '' ? undefined : e.target.value === 'true' 
              }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">全部状态</option>
              <option value="true">已启用</option>
              <option value="false">已禁用</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">执行状态</label>
            <select
              value={filters.last_run_status}
              onChange={(e) => setFilters(prev => ({ ...prev, last_run_status: e.target.value }))}
              className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">全部状态</option>
              <option value="pending">待执行</option>
              <option value="running">运行中</option>
              <option value="completed">已完成</option>
              <option value="failed">失败</option>
            </select>
          </div>
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="mb-4 rounded-md bg-red-50 p-4">
          <div className="flex">
            <XCircle className="h-5 w-5 text-red-400 flex-shrink-0" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">操作失败</h3>
              <p className="mt-1 text-sm text-red-700">{error}</p>
            </div>
            <div className="ml-auto">
              <button
                onClick={() => setError(null)}
                className="text-red-400 hover:text-red-600"
              >
                <XCircle className="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 任务列表 */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        {loading ? (
          <div className="p-12 text-center">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            <p className="mt-2 text-gray-500">加载中...</p>
          </div>
        ) : tasks.length === 0 ? (
          <div className="p-12 text-center">
            <Calendar className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">暂无定时任务</h3>
            <p className="mt-1 text-sm text-gray-500">开始创建你的第一个自动化报表任务</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    任务信息
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    报表类型
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    执行计划
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    执行状态
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    下次执行
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    操作
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {tasks.map((task) => (
                  <tr key={task.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div>
                        <div className="flex items-center">
                          <div className="text-sm font-medium text-gray-900">
                            {task.task_name}
                          </div>
                          <div className="ml-2">
                            {task.is_enabled ? (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                                启用
                              </span>
                            ) : (
                              <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                禁用
                              </span>
                            )}
                          </div>
                        </div>
                        <div className="text-sm text-gray-500">{task.task_description}</div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                      {getReportTypeName(task.report_type)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      <code className="bg-gray-100 px-2 py-1 rounded text-xs">
                        {task.cron_expression}
                      </code>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        {getStatusIcon(task.last_run_status)}
                        <span className="ml-2 text-sm text-gray-900">
                          {getStatusText(task.last_run_status)}
                        </span>
                      </div>
                      {task.last_run_time && (
                        <div className="text-xs text-gray-500">
                          {formatDateTime(task.last_run_time)}
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatDateTime(task.next_run_time)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex items-center space-x-2 justify-end">
                        <button
                          onClick={() => executeTask(task.id)}
                          disabled={executing.has(task.id)}
                          className="text-blue-600 hover:text-blue-900 disabled:opacity-50"
                          title="立即执行"
                        >
                          <Play className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => toggleTask(task.id, !task.is_enabled)}
                          className="text-yellow-600 hover:text-yellow-900"
                          title={task.is_enabled ? '禁用' : '启用'}
                        >
                          {task.is_enabled ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                        </button>
                        <button
                          onClick={() => {
                            setEditingTask(task);
                            setShowEditModal(true);
                          }}
                          className="text-indigo-600 hover:text-indigo-900"
                          title="编辑"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => deleteTask(task.id)}
                          className="text-red-600 hover:text-red-900"
                          title="删除"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 分页 */}
      {total > pageSize && (
        <div className="mt-6 flex items-center justify-between">
          <div className="text-sm text-gray-500">
            显示第 {Math.min((page - 1) * pageSize + 1, total)} - {Math.min(page * pageSize, total)} 条，
            共 {total} 条记录
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => setPage(prev => Math.max(1, prev - 1))}
              disabled={page === 1}
              className="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
            >
              上一页
            </button>
            <button
              onClick={() => setPage(prev => prev + 1)}
              disabled={page * pageSize >= total}
              className="px-3 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
            >
              下一页
            </button>
          </div>
        </div>
      )}

      {/* 这里应该添加创建和编辑任务的模态框组件 */}
      {/* CreateTaskModal 和 EditTaskModal 组件由于篇幅原因省略 */}
      {/* 在实际实现中需要添加这些组件 */}
    </div>
  );
};

export default ScheduledTaskPage;