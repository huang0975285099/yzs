/**
 * 测试 Mock 数据
 */

/**
 * 用户 Mock 数据
 */
const mockUsers = {
  admin: {
    id: 1,
    username: 'admin',
    name: '管理员',
    role: 'admin',
    token: 'mock-admin-token-12345',
  },

  operator: {
    id: 2,
    username: 'operator',
    name: '操作员',
    role: 'operator',
    token: 'mock-operator-token-12345',
  },

  inspector: {
    id: 3,
    username: 'inspector',
    name: '质检员',
    role: 'inspector',
    token: 'mock-inspector-token-12345',
  },

  statistician: {
    id: 4,
    username: 'statistician',
    name: '统计员',
    role: 'statistician',
    token: 'mock-statistician-token-12345',
  },
}

/**
 * 订单 Mock 数据
 */
const mockOrders = {
  empty: {
    data: [],
    total: 0,
    page: 1,
    pageSize: 10,
  },

  single: {
    data: [
      {
        id: 1,
        orderNo: 'ORD20240101001',
        status: 'pending',
        createdAt: '2024-01-01 10:00:00',
        operator: 'operator1',
      },
    ],
    total: 1,
    page: 1,
    pageSize: 10,
  },

  list: {
    data: [
      {
        id: 1,
        orderNo: 'ORD20240101001',
        status: 'pending',
        createdAt: '2024-01-01 10:00:00',
        operator: 'operator1',
      },
      {
        id: 2,
        orderNo: 'ORD20240101002',
        status: 'processing',
        createdAt: '2024-01-01 11:00:00',
        operator: 'operator2',
      },
      {
        id: 3,
        orderNo: 'ORD20240101003',
        status: 'completed',
        createdAt: '2024-01-01 12:00:00',
        operator: 'operator1',
      },
    ],
    total: 3,
    page: 1,
    pageSize: 10,
  },

  largeList: generateOrders(100),
}

/**
 * 生成订单数据
 * @param {number} count - 数量
 * @returns {object} 订单列表
 */
function generateOrders(count) {
  const statuses = ['pending', 'processing', 'completed', 'cancelled']
  const data = []

  for (let i = 1; i <= count; i++) {
    data.push({
      id: i,
      orderNo: `ORD${String(i).padStart(12, '0')}`,
      status: statuses[Math.floor(Math.random() * statuses.length)],
      createdAt: new Date(2024, 0, 1 + Math.floor(i / 10)).toISOString(),
      operator: `operator${Math.floor(Math.random() * 5) + 1}`,
    })
  }

  return {
    data,
    total: count,
    page: 1,
    pageSize: 10,
  }
}

/**
 * 仪表盘统计 Mock 数据
 */
const mockDashboardStats = {
  todayOrders: 156,
  pendingOrders: 23,
  completedOrders: 120,
  operatorsOnline: 8,
  successRate: 0.92,
}

/**
 * API 响应 Mock
 */
const apiResponses = {
  success: { success: true, message: '操作成功' },
  error: { success: false, message: '操作失败' },
  unauthorized: { success: false, message: '未授权', code: 401 },
  forbidden: { success: false, message: '无权限', code: 403 },
  notFound: { success: false, message: '资源不存在', code: 404 },
  serverError: { success: false, message: '服务器错误', code: 500 },
}

/**
 * 登录成功响应
 * @param {object} user - 用户对象
 * @returns {object} 登录响应
 */
function loginSuccessResponse(user) {
  return {
    success: true,
    data: {
      token: user.token,
      user: {
        id: user.id,
        username: user.username,
        name: user.name,
        role: user.role,
      },
    },
  }
}

/**
 * 登录失败响应
 */
const loginFailedResponse = {
  success: false,
  message: '用户名或密码错误',
}

module.exports = {
  mockUsers,
  mockOrders,
  mockDashboardStats,
  apiResponses,
  loginSuccessResponse,
  loginFailedResponse,
  generateOrders,
}